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

package dat

import (
	"context"
	"fmt"
	"time"

	jaeger "github.com/topfreegames/extensions/jaeger/dat"
	"gopkg.in/mgutz/dat.v2/dat"
)

// Execer executes queries against a database.
type Execer struct {
	ctx context.Context
	ex  dat.Execer
}

// NewExecer creates a new instance of Execer.
func NewExecer(execer dat.Execer) *Execer {
	return &Execer{
		ctx: nil,
		ex:  execer,
	}
}

// WithContext creates a shallow copy that uses the given context
func (jex *Execer) WithContext(ctx context.Context) *Execer {
	return &Execer{
		ctx: ctx,
		ex:  jex.ex,
	}
}

// Cache caches the results of queries for Select and SelectDoc.
func (jex *Execer) Cache(id string, ttl time.Duration, invalidate bool) dat.Execer {
	return jex.ex.Cache(id, ttl, invalidate)
}

// Timeout sets the timeout for current query.
func (jex *Execer) Timeout(timeout time.Duration) dat.Execer {
	return jex.ex.Timeout(timeout)
}

// Interpolate tells the associated builder to interpolate itself.
func (jex *Execer) Interpolate() (string, []interface{}, error) {
	return jex.ex.Interpolate()
}

// Exec executes a builder's query.
func (jex *Execer) Exec() (*dat.Result, error) {
	var result *dat.Result
	fullSQL, _, err := jex.ex.Interpolate()
	if err != nil {
		return result, err
	}

	jaeger.Trace(jex.ctx, fullSQL, func() error {
		result, err = jex.ex.Exec()
		return err
	})

	return result, err
}

// QueryScalar executes builder's query and scans returned row into destinations.
func (jex *Execer) QueryScalar(destinations ...interface{}) error {
	fullSQL, _, err := jex.ex.Interpolate()
	if err != nil {
		return err
	}

	jaeger.Trace(jex.ctx, fullSQL, func() error {
		err = jex.ex.QueryScalar(destinations...)
		return err
	})

	return err
}

// QuerySlice executes builder's query and builds a slice of values from each row, where
// each row only has one column.
func (jex *Execer) QuerySlice(dest interface{}) error {
	fullSQL, _, err := jex.ex.Interpolate()
	if err != nil {
		return err
	}

	jaeger.Trace(jex.ctx, fullSQL, func() error {
		err = jex.ex.QuerySlice(dest)
		return err
	})

	return err
}

// QueryStruct executes builders' query and scans the result row into dest.
func (jex *Execer) QueryStruct(dest interface{}) error {
	fullSQL, _, err := jex.ex.Interpolate()
	if err != nil {
		return err
	}

	jaeger.Trace(jex.ctx, fullSQL, func() error {
		err = jex.ex.QueryStruct(dest)
		return err
	})

	return err
}

// QueryStructs executes builders' query and scans each row as an item in a slice of structs.
func (jex *Execer) QueryStructs(dest interface{}) error {
	fullSQL, _, err := jex.ex.Interpolate()
	if err != nil {
		return err
	}

	jaeger.Trace(jex.ctx, fullSQL, func() error {
		err = jex.ex.QueryStructs(dest)
		return err
	})

	return err
}

// QueryObject wraps the builder's query within a `to_json` then executes and unmarshals
// the result into dest.
func (jex *Execer) QueryObject(dest interface{}) error {
	fullSQL, _, err := jex.ex.Interpolate()
	if err != nil {
		return err
	}

	jaeger.Trace(jex.ctx, fullSQL, func() error {
		err = jex.ex.QueryObject(dest)
		return err
	})

	return err
}

// QueryJSON wraps the builder's query within a `to_json` then executes and returns
// the JSON []byte representation.
func (jex *Execer) QueryJSON() ([]byte, error) {
	var result []byte
	fullSQL, _, err := jex.ex.Interpolate()
	if err != nil {
		return result, err
	}

	jaeger.Trace(jex.ctx, fullSQL, func() error {
		result, err = jex.ex.QueryJSON()
		return err
	})

	return result, err
}
