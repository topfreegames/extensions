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

package db

import (
	"context"
	"errors"
	"io"

	pg "github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	jaeger "github.com/topfreegames/extensions/jaeger/pg"
	"github.com/topfreegames/extensions/pg/interfaces"
)

// DB implements the orm.DB interface with a few tweaks for tracing.
type DB struct {
	inner interfaces.DB
	tx    interfaces.Tx
}

func (db *DB) Select(model interface{}) error {
	return db.inner.Select(model)
}

func (db *DB) Insert(model ...interface{}) error {
	return db.inner.Insert(model...)
}

func (db *DB) Update(model interface{}) error {
	return db.inner.Update(model)
}

func (db *DB) Delete(model interface{}) error {
	return db.inner.Delete(model)
}

func (db *DB) CopyFrom(r io.Reader, query interface{}, params ...interface{}) (orm.Result, error) {
	return db.inner.CopyFrom(r, query, params...)
}

func (db *DB) CopyTo(w io.Writer, query interface{}, params ...interface{}) (orm.Result, error) {
	return db.inner.CopyTo(w, query, params...)
}

func (db *DB) FormatQuery(b []byte, query string, params ...interface{}) []byte {
	return db.inner.FormatQuery(b, query, params...)
}

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

func (db *DB) Model(model ...interface{}) *orm.Query {
	return db.inner.Model(model...).DB(db)
}

func (db *DB) Close() error {
	return db.inner.Close()
}

func (db *DB) Begin() (*pg.Tx, error) {
	var tx *pg.Tx
	var err error
	jaeger.Trace(db.inner.Context(), "BEGIN", func() error {
		tx, err = db.inner.Begin()
		return err
	})
	return tx, err
}

func (db *DB) WithContext(ctx context.Context) *pg.DB {
	return db.inner.WithContext(ctx)
}

func (db *DB) Context() context.Context {
	return db.inner.Context()
}

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

func WithContext(ctx context.Context, db interfaces.DB) interfaces.DB {
	return &DB{
		inner: db.WithContext(ctx),
	}
}

func Begin(db interfaces.DB) (interfaces.DB, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &DB{tx: tx, inner: db}, nil
}

func Rollback(db interfaces.DB) error {
	if v, ok := db.(interface{ Rollback() error }); ok {
		return v.Rollback()
	}
	return errors.New("db does not implement rollback")
}

func Commit(db interfaces.DB) error {
	if v, ok := db.(interface{ Commit() error }); ok {
		return v.Commit()
	}
	return errors.New("db does not implement commit")
}
