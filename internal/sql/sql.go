// Copyright © 2021 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"

	"emperror.dev/errors"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

func init() {
	sql.Register("postgresql", stdlib.GetDefaultDriver())
	rand.Seed(time.Now().UnixNano())
}

type Client interface {
	RunQuery(log.Logger) (string, error)
	GetDriver() string
}

type client struct {
	driver string
	dsn    string
	query  string

	sqlQueryRepeatCount    int
	sqlQueryRepeatCountMax int

	conn *sql.DB

	dialerFunc DialerFunc
}

type DialerFunc func(ctx context.Context, network string, address string) (net.Conn, error)

type SQLClientOption func(*client)

func WithDialerFunc(dialerFunc DialerFunc) SQLClientOption {
	return func(f *client) {
		f.dialerFunc = dialerFunc
	}
}

func NewClient(dsn, query string, sqlQueryRepeatCount, sqlQueryRepeatCountMax int, opts ...SQLClientOption) (*client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	client := &client{
		driver:                 u.Scheme,
		query:                  query,
		sqlQueryRepeatCount:    sqlQueryRepeatCount,
		sqlQueryRepeatCountMax: sqlQueryRepeatCountMax,
		dialerFunc:             (&net.Dialer{}).DialContext,
	}

	switch client.driver {
	case "postgresql":
		cc, err := pgx.ParseConfig(dsn)
		if err != nil {
			return nil, err
		}
		cc.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return client.dialerFunc(ctx, network, addr)
		}
		dsn = stdlib.RegisterConnConfig(cc)
		fmt.Println(dsn)
	case "mysql":
		mysql.RegisterDialContext("custom-dialer", func(ctx context.Context, addr string) (net.Conn, error) {
			return client.dialerFunc(ctx, "tcp", addr)
		})

		u.Scheme = ""
		u.Host = fmt.Sprintf("custom-dialer(%s)", u.Host)
		if l := u.String(); len(l) > 3 {
			dsn = l[2:]
		}
	default:
		return nil, errors.Errorf("invalid SQL driver: '%s'", u.Scheme)
	}

	for _, o := range opts {
		o(client)
	}

	client.dsn = dsn
	if sqlQueryRepeatCountMax < sqlQueryRepeatCount {
		client.sqlQueryRepeatCountMax = sqlQueryRepeatCount
	}

	return client, nil
}

func (c *client) GetDriver() string {
	return c.driver
}

func (c *client) RunQuery(logger log.Logger) (string, error) {
	var err error

	if c.conn == nil {
		c.conn, err = sql.Open(c.driver, c.dsn)
		if err != nil {
			return c.query, fmt.Errorf("unable to connect to database: %v", err)
		}
		c.conn.SetMaxOpenConns(50)
		c.conn.SetMaxIdleConns(25)
	}

	var count int
	if c.sqlQueryRepeatCountMax > c.sqlQueryRepeatCount {
		count = rand.Intn(c.sqlQueryRepeatCountMax-c.sqlQueryRepeatCount+1) + c.sqlQueryRepeatCount
	}

	logger.WithFields(log.Fields{
		"query": c.query,
		"count": count,
	}).Info("outgoing query")

	for i := 0; i < count; i++ {
		rows, err := c.conn.Query(c.query)
		if err != nil {
			return c.query, fmt.Errorf("query failed: %v", err)
		}

		if err := rows.Close(); err != nil {
			return c.query, fmt.Errorf("query failed: %v", err)
		}
	}

	return c.query, nil
}
