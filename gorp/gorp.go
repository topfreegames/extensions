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
	"reflect"
	"regexp"

	"github.com/go-gorp/gorp/v3"
	"github.com/topfreegames/extensions/v9/gorp/interfaces"
	tgorp "github.com/topfreegames/extensions/v9/tracing/gorp"
)

// New wraps and instruments an existing connection to a database
func New(inner *gorp.DbMap, name string) *Database {
	return &Database{
		executor: &executor{nil, inner, name},
	}
}

// Database wraps a gorp database connection
type Database struct {
	*executor
}

// WithContext creates a shallow copy that uses the given context
func (d *Database) WithContext(ctx context.Context) gorp.SqlExecutor {
	return &Database{
		executor: d.executor.WithContext(ctx).(*executor),
	}
}

// Begin starts a gorp Transaction
func (d *Database) Begin() (interfaces.Transaction, error) {
	var inner *gorp.Transaction
	var err error

	tgorp.Trace(d.ctx, d.executor.name, "BEGIN", func() error {
		inner, err = d.Inner().Begin()
		return err
	})

	if err != nil {
		return nil, err
	}

	result := &Transaction{
		executor: &executor{d.ctx, inner, d.executor.name},
	}

	return result, err
}

// Close closes the database, releasing any open resources
func (d *Database) Close() error {
	return d.Inner().Db.Close()
}

// Inner returns the wrapped gorp database
func (d *Database) Inner() *gorp.DbMap {
	return d.executor.inner.(*gorp.DbMap)
}

// Transaction wraps a gorp transaction
type Transaction struct {
	*executor
}

// WithContext creates a shallow copy that uses the given context
func (t *Transaction) WithContext(ctx context.Context) gorp.SqlExecutor {
	return &Transaction{
		executor: t.executor.WithContext(ctx).(*executor),
	}
}

// Commit commits the underlying database transaction
func (t *Transaction) Commit() error {
	var err error

	tgorp.Trace(t.ctx, t.executor.name, "COMMIT", func() error {
		err = t.Inner().Commit()
		return err
	})

	return err
}

// Rollback rolls back the underlying database transaction
func (t *Transaction) Rollback() error {
	var err error

	tgorp.Trace(t.ctx, t.executor.name, "ROLLBACK", func() error {
		err = t.Inner().Rollback()
		return err
	})

	return err
}

// Inner returns the wrapped gorp transaction
func (t *Transaction) Inner() *gorp.Transaction {
	return t.executor.inner.(*gorp.Transaction)
}

type executor struct {
	ctx   context.Context
	inner gorp.SqlExecutor
	name  string
}

// WithContext creates a shallow copy that uses the given context
func (e *executor) WithContext(ctx context.Context) gorp.SqlExecutor {
	return &executor{
		inner: e.inner.WithContext(ctx),
		ctx:   ctx,
		name:  e.name,
	}
}

// Get runs a SQL SELECT to fetch a single row from the table based on the primary key(s)
func (e *executor) Get(i interface{}, keys ...interface{}) (interface{}, error) {
	var result interface{}
	var err error

	query := fmt.Sprintf("SELECT <type:%T keys:%v>", i, keys)

	tgorp.Trace(e.ctx, e.name, query, func() error {
		result, err = e.inner.Get(i, keys...)
		return err
	})

	return result, err
}

// Insert runs a SQL INSERT statement for each element in list
func (e *executor) Insert(list ...interface{}) error {
	var err error

	types := types(list)
	query := fmt.Sprintf("MULTI-INSERT <objects:%v>", types)

	tgorp.Trace(e.ctx, e.name, query, func() error {
		err = e.inner.Insert(list...)
		return err
	})

	return err
}

// Update runs a SQL UPDATE statement for each element in list
func (e *executor) Update(list ...interface{}) (int64, error) {
	var result int64
	var err error

	types := types(list)
	query := fmt.Sprintf("MULTI-UPDATE <objects:%v>", types)

	tgorp.Trace(e.ctx, e.name, query, func() error {
		result, err = e.inner.Update(list...)
		return err
	})

	return result, err
}

// Delete runs a SQL DELETE statement for each element in list
func (e *executor) Delete(list ...interface{}) (int64, error) {
	var result int64
	var err error

	types := types(list)
	query := fmt.Sprintf("MULTI-DELETE <objects:%v>", types)

	tgorp.Trace(e.ctx, e.name, query, func() error {
		result, err = e.inner.Delete(list...)
		return err
	})

	return result, err
}

// Exec runs an arbitrary SQL statement
func (e *executor) Exec(query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result, err = e.inner.Exec(query, args...)
		return err
	})

	return result, err
}

// Select runs an arbitrary SQL query, binding the columns in the result to fields on the struct specified by i
func (e *executor) Select(i interface{}, query string, args ...interface{}) ([]interface{}, error) {
	var result []interface{}
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result, err = e.inner.Select(i, query, args...)
		return err
	})

	return result, err
}

// SelectInt is a convenience wrapper around the gorp.SelectInt function
func (e *executor) SelectInt(query string, args ...interface{}) (int64, error) {
	var result int64
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result, err = e.inner.SelectInt(query, args...)
		return err
	})

	return result, err
}

// SelectNullInt is a convenience wrapper around the gorp.SelectNullInt function
func (e *executor) SelectNullInt(query string, args ...interface{}) (sql.NullInt64, error) {
	var result sql.NullInt64
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result, err = e.inner.SelectNullInt(query, args...)
		return err
	})

	return result, err
}

// SelectFloat is a convenience wrapper around the gorp.SelectFloat function
func (e *executor) SelectFloat(query string, args ...interface{}) (float64, error) {
	var result float64
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result, err = e.inner.SelectFloat(query, args...)
		return err
	})

	return result, err
}

// SelectNullFloat is a convenience wrapper around the gorp.SelectNullFloat function
func (e *executor) SelectNullFloat(query string, args ...interface{}) (sql.NullFloat64, error) {
	var result sql.NullFloat64
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result, err = e.inner.SelectNullFloat(query, args...)
		return err
	})

	return result, err
}

// SelectStr is a convenience wrapper around the gorp.SelectStr function
func (e *executor) SelectStr(query string, args ...interface{}) (string, error) {
	var result string
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result, err = e.inner.SelectStr(query, args...)
		return err
	})

	return result, err
}

// SelectNullStr is a convenience wrapper around the gorp.SelectNullStr function
func (e *executor) SelectNullStr(query string, args ...interface{}) (sql.NullString, error) {
	var result sql.NullString
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result, err = e.inner.SelectNullStr(query, args...)
		return err
	})

	return result, err
}

// SelectOne is a convenience wrapper around the gorp.SelectOne function
func (e *executor) SelectOne(holder interface{}, query string, args ...interface{}) error {
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		err = e.inner.SelectOne(holder, query, args...)
		return err
	})

	return err
}

func (e *executor) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var result *sql.Rows
	var err error

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result, err = e.inner.Query(query, args...)
		return err
	})

	return result, err
}

func (e *executor) QueryRow(query string, args ...interface{}) *sql.Row {
	var result *sql.Row

	tgorp.Trace(e.ctx, e.name, format(query, args), func() error {
		result = e.inner.QueryRow(query, args...)
		return nil
	})

	return result
}

func format(query string, args []interface{}) string {
	re := regexp.MustCompile("\\$(\\d+)")
	template := re.ReplaceAllString(query, "%[$1]v")
	return fmt.Sprintf(template, args...)
}

func types(list []interface{}) []reflect.Type {
	var result []reflect.Type
	for _, val := range list {
		result = append(result, reflect.TypeOf(val))
	}
	return result
}
