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
	"net/url"

	"emperror.dev/errors"
	// mysql driver
	_ "github.com/go-mysql-org/go-mysql/driver"
	"github.com/jackc/pgx/v4/stdlib"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

func init() {
	sql.Register("postgresql", stdlib.GetDefaultDriver())
}

type Client struct {
	driver string
	dsn    string
	query  string
	repeat int

	conn *sql.DB
}

func NewClient(dsn, query string, repeat int) (*Client, error) {
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

	return &Client{
		driver: driver,
		dsn:    dsn,
		query:  query,
		repeat: repeat,
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
			return c.query, fmt.Errorf("Unable to connect to database: %v\n", err)
		}
		c.conn.SetMaxOpenConns(50)
		c.conn.SetMaxIdleConns(25)
	}

	for i := 0; i < c.repeat; i++ {
		rows, err := c.conn.Query(c.query)
		rows.Close()

		if err != nil {
			return c.query, fmt.Errorf("query failed: %v\n", err)
		}
		logger.WithFields(log.Fields{
			"query": c.query,
		}).Info("outgoing query")
	}

	return c.query, nil
}
