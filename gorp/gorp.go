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
	"fmt"
	"regexp"

	"github.com/go-gorp/gorp"
	jgorp "github.com/topfreegames/extensions/jaeger/gorp"
)

func format(query string, args []interface{}) string {
	re := regexp.MustCompile("\\$(\\d+)")
	template := re.ReplaceAllString(query, "%[$1]v")
	return fmt.Sprintf(template, args...)
}

type DbMap struct {
	*gorp.DbMap
	ctx context.Context
}

func (m *DbMap) WithContext(ctx context.Context) gorp.SqlExecutor {
	return &DbMap{
		m.DbMap.WithContext(ctx).(*gorp.DbMap),
		ctx,
	}
}

func (m *DbMap) Exec(query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	jgorp.Trace(m.ctx, format(query, args), func() error {
		result, err = m.DbMap.Exec(query, args...)
		return err
	})

	return result, err
}

func (m *DbMap) Begin() (*Transaction, error) {
	var inner *gorp.Transaction
	var err error

	jgorp.Trace(m.ctx, "BEGIN", func() error {
		inner, err = m.DbMap.Begin()
		return err
	})

	if err != nil {
		return nil, err
	}

	result := &Transaction{
		inner.WithContext(m.ctx).(*gorp.Transaction),
		m.ctx,
	}

	return result, err
}

func (m *DbMap) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var result *sql.Rows
	var err error

	jgorp.Trace(m.ctx, format(query, args), func() error {
		result, err = m.DbMap.Query(query, args...)
		return err
	})

	return result, err
}

func (m *DbMap) QueryRow(query string, args ...interface{}) *sql.Row {
	var result *sql.Row

	jgorp.Trace(m.ctx, format(query, args), func() error {
		result = m.DbMap.QueryRow(query, args...)
		return nil
	})

	return result
}

type Transaction struct {
	*gorp.Transaction
	ctx context.Context
}

func (t *Transaction) WithContext(ctx context.Context) gorp.SqlExecutor {
	return &Transaction{
		t.Transaction.WithContext(ctx).(*gorp.Transaction),
		ctx,
	}
}

func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	jgorp.Trace(t.ctx, format(query, args), func() error {
		result, err = t.Transaction.Exec(query, args...)
		return err
	})

	return result, err
}

func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var result *sql.Rows
	var err error

	jgorp.Trace(t.ctx, format(query, args), func() error {
		result, err = t.Transaction.Query(query, args...)
		return err
	})

	return result, err
}

func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	var result *sql.Row

	jgorp.Trace(t.ctx, format(query, args), func() error {
		result = t.Transaction.QueryRow(query, args...)
		return nil
	})

	return result
}

func (t *Transaction) Commit() error {
	var err error

	jgorp.Trace(t.ctx, "COMMIT", func() error {
		err = t.Transaction.Commit()
		return err
	})

	return err
}

func (t *Transaction) Rollback() error {
	var err error

	jgorp.Trace(t.ctx, "ROLLBACK", func() error {
		err = t.Transaction.Rollback()
		return err
	})

	return err
}
