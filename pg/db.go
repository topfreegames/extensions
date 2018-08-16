/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
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

package pg

import (
	"context"
	"errors"
	"io"

	pg "github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	jaeger "github.com/topfreegames/extensions/jaeger/pg"
	"github.com/topfreegames/extensions/pg/interfaces"
)

// DB implements the orm.DB interface with a few tweaks for tracing
type DB struct {
	inner interfaces.DB
	tx    interfaces.Tx
}

// Select calls inner db select
func (db *DB) Select(model interface{}) error {
	return db.inner.Select(model)
}

// Insert calls inner db insert
func (db *DB) Insert(model ...interface{}) error {
	return db.inner.Insert(model...)
}

// Update calls inner db update
func (db *DB) Update(model interface{}) error {
	return db.inner.Update(model)
}

// Delete calls inner db delete
func (db *DB) Delete(model interface{}) error {
	return db.inner.Delete(model)
}

// CopyFrom calls inner db copyFrom
func (db *DB) CopyFrom(r io.Reader, query interface{}, params ...interface{}) (orm.Result, error) {
	return db.inner.CopyFrom(r, query, params...)
}

// CopyTo calls inner db copyTo
func (db *DB) CopyTo(w io.Writer, query interface{}, params ...interface{}) (orm.Result, error) {
	return db.inner.CopyTo(w, query, params...)
}

// FormatQuery calls inner db formatQuery
func (db *DB) FormatQuery(b []byte, query string, params ...interface{}) []byte {
	return db.inner.FormatQuery(b, query, params...)
}

// Close calls inner db close
func (db *DB) Close() error {
	return db.inner.Close()
}

// WithContext calls inner db withContext
// Panics if a transaction is passed as argument
func (db *DB) WithContext(ctx context.Context) *pg.DB {
	if db.tx != nil {
		// is actually a transaction
		// panic because we can't return an error without breaking the interface contract
		panic("cannot call WithContext in a transaction")
	}
	return db.inner.WithContext(ctx)
}

// Context calls inner db context
func (db *DB) Context() context.Context {
	return db.inner.Context()
}

// Model calls inner db model and assigns to the returned query the current db
// This ensures calls to Model return DB structs instead of pg.DB
func (db *DB) Model(model ...interface{}) *orm.Query {
	return db.inner.Model(model...).DB(db)
}

// Exec wraps the inner db or tx Exec call in a jaeger trace
// If a tx is available it will be used
func (db *DB) Exec(query interface{}, params ...interface{}) (orm.Result, error) {
	var res orm.Result
	var err error
	jaeger.Trace(db.inner.Context(), query, func() error {
		if db.tx != nil {
			res, err = db.tx.Exec(query, params...)
		} else {
			res, err = db.inner.Exec(query, params...)
		}
		return err
	})
	return res, err
}

// ExecOne wraps the inner db or tx ExecOne call in a jaeger trace
// If a tx is available it will be used
func (db *DB) ExecOne(query interface{}, params ...interface{}) (orm.Result, error) {
	var res orm.Result
	var err error
	jaeger.Trace(db.inner.Context(), query, func() error {
		if db.tx != nil {
			res, err = db.tx.ExecOne(query, params...)
		} else {
			res, err = db.inner.ExecOne(query, params...)
		}
		return err
	})
	return res, err
}

// Query wraps the inner db or tx Query call in a jaeger trace
// If a tx is available it will be used
func (db *DB) Query(model, query interface{}, params ...interface{}) (orm.Result, error) {
	var res orm.Result
	var err error
	jaeger.Trace(db.inner.Context(), query, func() error {
		if db.tx != nil {
			res, err = db.tx.Query(model, query, params...)
		} else {
			res, err = db.inner.Query(model, query, params...)
		}
		return err
	})
	return res, err
}

// QueryOne wraps the inner db or tx QueryOne call in a jaeger trace
// If a tx is available it will be used
func (db *DB) QueryOne(model, query interface{}, params ...interface{}) (orm.Result, error) {
	var res orm.Result
	var err error
	jaeger.Trace(db.inner.Context(), query, func() error {
		if db.tx != nil {
			res, err = db.tx.QueryOne(model, query, params...)
		} else {
			res, err = db.inner.QueryOne(model, query, params...)
		}
		return err
	})
	return res, err
}

// Begin wraps the inner db Begin call in a jaeger trace
func (db *DB) Begin() (*pg.Tx, error) {
	var tx *pg.Tx
	var err error
	jaeger.Trace(db.inner.Context(), "BEGIN", func() error {
		tx, err = db.inner.Begin()
		return err
	})
	return tx, err
}

// Rollback wraps the inner db Rollback call in a jaeger trace
// It returns an error if no tx is available in the current db
func (db *DB) Rollback() error {
	var err error
	jaeger.Trace(db.inner.Context(), "ROLLBACK", func() error {
		if db.tx != nil {
			err = db.tx.Rollback()
		} else {
			err = errors.New("cannot rollback if no transaction")
		}
		return err
	})
	return err
}

// Commit wraps the inner db Commit call in a jaeger trace
// It returns an error if no tx is available in the current db
func (db *DB) Commit() error {
	var err error
	jaeger.Trace(db.inner.Context(), "COMMIT", func() error {
		if db.tx != nil {
			err = db.tx.Commit()
		} else {
			err = errors.New("cannot commit if no transaction")
		}
		return err
	})
	return err
}

// WithContext calls the given db WithContext method and returns a DB with the new pg.DB
// If a transaction is passed as argument it is just returned without setting a new ctx
func WithContext(ctx context.Context, db interfaces.DB) interfaces.DB {
	// the cases on if/else are actually transactions: noop
	if sdb, ok := db.(*DB); ok {
		if sdb.tx != nil {
			return db
		}
	} else if _, ok := db.(interface{ Rollback() error }); ok {
		return db
	}

	return &DB{
		inner: db.WithContext(ctx),
	}
}

// Begin calls the given db Begin method and returns a DB with the new pg.Tx
func Begin(db interfaces.DB) (interfaces.DB, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &DB{tx: tx, inner: db}, nil
}

// Rollback calls the given db Rollback method and returns an error if it does not implement it
func Rollback(db interfaces.DB) error {
	if v, ok := db.(interface{ Rollback() error }); ok {
		return v.Rollback()
	}
	return errors.New("db does not implement rollback")
}

// Commit calls the given db Commit method and returns an error if it does not implement it
func Commit(db interfaces.DB) error {
	if v, ok := db.(interface{ Commit() error }); ok {
		return v.Commit()
	}
	return errors.New("db does not implement commit")
}
