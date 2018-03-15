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
	"github.com/topfreegames/extensions/gorp/interfaces"
	jgorp "github.com/topfreegames/extensions/jaeger/gorp"
)

func New(Inner *gorp.DbMap) *Database {
	return &Database{
		executor: &executor{Inner, nil},
	}
}

type Database struct {
	*executor
}

func (d *Database) WithContext(ctx context.Context) gorp.SqlExecutor {
	return &Database{
		executor: d.executor.WithContext(ctx).(*executor),
	}
}

func (d *Database) Begin() (interfaces.Transaction, error) {
	var Inner *gorp.Transaction
	var err error

	jgorp.Trace(d.ctx, "BEGIN", func() error {
		Inner, err = d.Inner().Begin()
		return err
	})

	if err != nil {
		return nil, err
	}

	result := &Transaction{
		executor: &executor{Inner, d.ctx},
	}

	return result, err
}

func (d *Database) Close() error {
	return d.Inner().Db.Close()
}

func (d *Database) Inner() *gorp.DbMap {
	return d.executor.Inner.(*gorp.DbMap)
}

type Transaction struct {
	*executor
}

func (t *Transaction) WithContext(ctx context.Context) gorp.SqlExecutor {
	return &Transaction{
		executor: t.executor.WithContext(ctx).(*executor),
	}
}

func (t *Transaction) Commit() error {
	var err error

	jgorp.Trace(t.ctx, "COMMIT", func() error {
		err = t.Inner().Commit()
		return err
	})

	return err
}

func (t *Transaction) Rollback() error {
	var err error

	jgorp.Trace(t.ctx, "ROLLBACK", func() error {
		err = t.Inner().Rollback()
		return err
	})

	return err
}

func (t *Transaction) Inner() *gorp.Transaction {
	return t.executor.Inner.(*gorp.Transaction)
}

type executor struct {
	Inner gorp.SqlExecutor
	ctx   context.Context
}

func (e *executor) WithContext(ctx context.Context) gorp.SqlExecutor {
	return &executor{
		Inner: e.Inner.WithContext(ctx),
		ctx:   ctx,
	}
}

func (e *executor) Get(i interface{}, keys ...interface{}) (interface{}, error) {
	return e.Inner.Get(i, keys...)
}

func (e *executor) Insert(list ...interface{}) error {
	return e.Inner.Insert(list...)
}

func (e *executor) Update(list ...interface{}) (int64, error) {
	return e.Inner.Update(list...)
}

func (e *executor) Delete(list ...interface{}) (int64, error) {
	return e.Inner.Delete(list...)
}

func (e *executor) Exec(query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result, err = e.Inner.Exec(query, args...)
		return err
	})

	return result, err
}

func (e *executor) Select(i interface{}, query string, args ...interface{}) ([]interface{}, error) {
	var result []interface{}
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result, err = e.Inner.Select(i, query, args...)
		return err
	})

	return result, err
}

func (e *executor) SelectInt(query string, args ...interface{}) (int64, error) {
	var result int64
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result, err = e.Inner.SelectInt(query, args...)
		return err
	})

	return result, err
}

func (e *executor) SelectNullInt(query string, args ...interface{}) (sql.NullInt64, error) {
	var result sql.NullInt64
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result, err = e.Inner.SelectNullInt(query, args...)
		return err
	})

	return result, err
}

func (e *executor) SelectFloat(query string, args ...interface{}) (float64, error) {
	var result float64
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result, err = e.Inner.SelectFloat(query, args...)
		return err
	})

	return result, err
}

func (e *executor) SelectNullFloat(query string, args ...interface{}) (sql.NullFloat64, error) {
	var result sql.NullFloat64
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result, err = e.Inner.SelectNullFloat(query, args...)
		return err
	})

	return result, err
}

func (e *executor) SelectStr(query string, args ...interface{}) (string, error) {
	var result string
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result, err = e.Inner.SelectStr(query, args...)
		return err
	})

	return result, err
}

func (e *executor) SelectNullStr(query string, args ...interface{}) (sql.NullString, error) {
	var result sql.NullString
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result, err = e.Inner.SelectNullStr(query, args...)
		return err
	})

	return result, err
}

func (e *executor) SelectOne(holder interface{}, query string, args ...interface{}) error {
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		err = e.Inner.SelectOne(holder, query, args...)
		return err
	})

	return err
}

func (e *executor) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var result *sql.Rows
	var err error

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result, err = e.Inner.Query(query, args...)
		return err
	})

	return result, err
}

func (e *executor) QueryRow(query string, args ...interface{}) *sql.Row {
	var result *sql.Row

	jgorp.Trace(e.ctx, format(query, args), func() error {
		result = e.Inner.QueryRow(query, args...)
		return nil
	})

	return result
}

func format(query string, args []interface{}) string {
	re := regexp.MustCompile("\\$(\\d+)")
	template := re.ReplaceAllString(query, "%[$1]v")
	return fmt.Sprintf(template, args...)
}
