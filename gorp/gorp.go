/*
 * Copyright (c) 2018 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package gorp

import (
	"context"
	"database/sql"

	"github.com/go-gorp/gorp"
	jgorp "github.com/topfreegames/extensions/jaeger/gorp"
)

type DbMap struct {
	*gorp.DbMap
	ctx context.Context
}

func (m *DbMap) WithContext(ctx context.Context) gorp.SqlExecutor {
	copy := DbMap{m.DbMap, ctx}
	return copy.WithContext(ctx)
}

func (m *DbMap) Exec(query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	jgorp.Trace(m.ctx, query, func() error {
		result, err = m.DbMap.Exec(query, args)
		return err
	})

	return result, err
}

// func (m *DbMap) Prepare(query string) (*sql.Stmt, error) {
// 	return m.DbMap.Prepare(query)
// }

func (m *DbMap) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var result *sql.Rows
	var err error

	jgorp.Trace(m.ctx, query, func() error {
		result, err = m.DbMap.Query(query, args)
		return err
	})

	return result, err
}

func (m *DbMap) QueryRow(query string, args ...interface{}) *sql.Row {
	var result *sql.Row

	jgorp.Trace(m.ctx, query, func() error {
		result = m.DbMap.QueryRow(query, args)
		return nil
	})

	return result
}
