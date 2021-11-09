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
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"emperror.dev/errors"
	// mysql driver
	_ "github.com/go-mysql-org/go-mysql/driver"
	"github.com/jackc/pgx/v4/stdlib"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

func init() {
	sql.Register("postgresql", stdlib.GetDefaultDriver())
	rand.Seed(time.Now().UnixNano())
}

type Client struct {
	driver string
	dsn    string
	query  string

	sqlQueryRepeatCount    int
	sqlQueryRepeatCountMax int

	conn *sql.DB
}

func NewClient(dsn, query string, sqlQueryRepeatCount, sqlQueryRepeatCountMax int) (*Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	driver := u.Scheme

	switch driver {
	case "postgresql":
	case "mysql":
		u.Scheme = ""
		if l := u.String(); len(l) > 3 {
			dsn = l[2:]
		}
	default:
		return nil, errors.Errorf("invalid SQL driver: '%s'", u.Scheme)
	}

	if sqlQueryRepeatCountMax < sqlQueryRepeatCount {
		sqlQueryRepeatCountMax = sqlQueryRepeatCount
	}

	return &Client{
		driver:                 driver,
		dsn:                    dsn,
		query:                  query,
		sqlQueryRepeatCount:    sqlQueryRepeatCount,
		sqlQueryRepeatCountMax: sqlQueryRepeatCountMax,
	}, nil
}

func (c *Client) GetDriver() string {
	return c.driver
}

func (c *Client) RunQuery(logger log.Logger) (string, error) {
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
		rows.Close()
	}

	return c.query, nil
}
