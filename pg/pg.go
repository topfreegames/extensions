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
	"fmt"
	"io"
	"time"

	pg "github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"

	"github.com/spf13/viper"
	jaeger "github.com/topfreegames/extensions/jaeger/pg"
	"github.com/topfreegames/extensions/pg/interfaces"
)

// Client is the struct that connects to PostgreSQL
type Client struct {
	Config    *viper.Viper
	DB        interfaces.DB
	Options   *pg.Options
	TxWrapper interfaces.TxWrapper
	context   context.Context
}

// TxWrapper is the struct for the TxWrapper
type TxWrapper struct{}

// DbBegin is a wrapper for returning db transactions
func (t *TxWrapper) DbBegin(db interfaces.DB) (interfaces.Tx, error) {
	return db.Begin()
}

// NewClient creates a new client
func NewClient(prefix string, config *viper.Viper, pgOrNil interfaces.DB, txOrNil interfaces.TxWrapper) (*Client, error) {
	client := &Client{
		Config: config,
	}

	var db interfaces.DB
	if pgOrNil != nil {
		db = pgOrNil
	}
	var tx interfaces.TxWrapper
	if txOrNil != nil {
		tx = txOrNil
		client.TxWrapper = tx
	} else {
		client.TxWrapper = &TxWrapper{}
	}
	err := client.Connect(prefix, db)
	if err != nil {
		return nil, err
	}

	if pgOrNil == nil {
		timeout := config.GetInt(fmt.Sprintf("%s.connectionTimeout", prefix))
		err = client.WaitForConnection(timeout)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

// Connect to PG
func (c *Client) Connect(prefix string, db interfaces.DB) error {
	user := c.Config.GetString(fmt.Sprintf("%s.user", prefix))
	pass := c.Config.GetString(fmt.Sprintf("%s.pass", prefix))
	host := c.Config.GetString(fmt.Sprintf("%s.host", prefix))
	database := c.Config.GetString(fmt.Sprintf("%s.database", prefix))
	port := c.Config.GetInt(fmt.Sprintf("%s.port", prefix))
	poolSize := c.Config.GetInt(fmt.Sprintf("%s.poolSize", prefix))
	maxRetries := c.Config.GetInt(fmt.Sprintf("%s.maxRetries", prefix))

	c.Options = &pg.Options{
		Addr:       fmt.Sprintf("%s:%d", host, port),
		User:       user,
		Password:   pass,
		Database:   database,
		PoolSize:   poolSize,
		MaxRetries: maxRetries,
	}

	if db == nil {
		c.DB = &DB{inner: pg.Connect(c.Options)}
	} else {
		c.DB = &DB{inner: db}
	}

	return nil
}

// IsConnected determines if the client is connected to PG
func (c *Client) IsConnected() bool {
	res, err := c.DB.Exec("select 1")
	if err != nil {
		return false
	}
	return res.RowsReturned() == 1
}

// Close the connections to PG
func (c *Client) Close() error {
	err := c.DB.Close()
	if err != nil {
		return err
	}
	return nil
}

// WaitForConnection loops until PG is connected
func (c *Client) WaitForConnection(timeout int) error {
	t := time.Duration(timeout) * time.Second
	timeoutTimer := time.NewTimer(t)
	defer timeoutTimer.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			return fmt.Errorf("timed out waiting for PostgreSQL to connect")
		case <-ticker.C:
			if c.IsConnected() {
				return nil
			}
		}
	}
}

//Cleanup closes PG connection
func (c *Client) Cleanup() error {
	err := c.Close()
	return err
}

func (c *Client) WithContext(ctx context.Context) *Client {
	return &Client{
		Config:    c.Config,
		DB:        c.DB,
		Options:   c.Options,
		TxWrapper: c.TxWrapper,
		context:   ctx,
	}
}

func (c *Client) Context() context.Context {
	return c.context
}

// TODO camila probably move this to another file
// DB implements the orm.DB interface
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
