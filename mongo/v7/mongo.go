/*
 * Copyright (c) 2017 TFG Co <backend@tfgco.com>
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

package v7

import (
	"context"
	"fmt"
	"strings"

	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	tracing "github.com/topfreegames/extensions/v9/tracing/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mongo holds the mongo db and connection.
// Implements MongoDB interface using the mongo-driver package
type mongoImpl struct {
	ctx    context.Context
	client *mongo.Client
	db     *mongo.Database
}

// NewMongo return Mongo instance with completed fields
func newMongo(client *mongo.Client, db *mongo.Database) interfaces.MongoDB {
	return &mongoImpl{
		ctx:    nil,
		client: client,
		db:     db,
	}
}

// WithContext creates a shallow copy that uses the given context
func (m *mongoImpl) WithContext(ctx context.Context) interfaces.MongoDB {
	return &mongoImpl{
		ctx:    ctx,
		client: m.client,
	}
}

// Run executes run command on db
func (m *mongoImpl) Run(cmd interface{}, result interface{}) error {
	var err error

	database := m.db.Name()
	args := formatArgs(cmd)

	tracing.Trace(m.ctx, database, database, "runCommand", args, func() error {
		err = m.db.RunCommand(m.ctx, cmd).Decode(result)
		return err
	})

	return err
}

// C returns the collection from database and a session
// This session needs to be closed afterward
func (m *mongoImpl) C(name string) (interfaces.Collection, interfaces.Session) {
	session, err := m.client.StartSession()
	c := &Collection{
		ctx:        m.ctx,
		collection: m.db.With(session).C(name),
	}
	return c, session
}

// Close closes mongo session
func (m *mongoImpl) Close() {
	m.session.Close()
}

// Collection holds a mongo collection and implements Collection interface
type Collection struct {
	ctx        context.Context
	collection *mgo.Collection
}

// Find executes a find query on Mongo
func (c *Collection) Find(query interface{}) interfaces.Query {
	return &Query{
		ctx:   c.ctx,
		query: c.collection.Find(query),

		database: c.collection.Database.Name,
		prefix:   c.collection.FullName,
		args:     formatArgs(query),
	}
}

// FindId is a conveniene method to execute a find by id query on Mongo
func (c *Collection) FindId(id interface{}) interfaces.Query {
	return &Query{
		ctx:   c.ctx,
		query: c.collection.FindId(id),

		database: c.collection.Database.Name,
		prefix:   c.collection.FullName,
		args:     formatArgs(bson.D{{Name: "_id", Value: id}}),
	}
}

// Pipe calls mongo collection Pipe
func (c *Collection) Pipe(pipeline interface{}) interfaces.Pipe {
	return &Pipe{
		ctx:  c.ctx,
		pipe: c.collection.Pipe(pipeline),

		database: c.collection.Database.Name,
		prefix:   c.collection.FullName,
		args:     formatArgs(pipeline),
	}
}

// Insert calls mongo collection Insert
func (c *Collection) Insert(docs ...interface{}) error {
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(docs)

	tracing.Trace(c.ctx, database, collection, "insert", args, func() error {
		err = c.collection.Insert(docs...)
		return err
	})

	return err
}

// UpsertId calls mongo collection UpsertId
func (c *Collection) UpsertId(id interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	var result *mgo.ChangeInfo
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(bson.D{{Name: "_id", Value: id}}, update)

	tracing.Trace(c.ctx, database, collection, "updateOne", args, func() error {
		result, err = c.collection.UpsertId(id, update)
		return err
	})

	return result, err
}

// Upsert calls mongo collection Upsert
func (c *Collection) Upsert(selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	var result *mgo.ChangeInfo
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(selector, update)

	tracing.Trace(c.ctx, database, collection, "updateOne", args, func() error {
		result, err = c.collection.Upsert(selector, update)
		return err
	})

	return result, err
}

// RemoveId calls mongo collection RemoveId
func (c *Collection) RemoveId(id interface{}) error {
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(bson.D{{Name: "_id", Value: id}})

	tracing.Trace(c.ctx, database, collection, "remove", args, func() error {
		err = c.collection.RemoveId(id)
		return err
	})

	return err
}

// Remove calls mongo collection Remove
func (c *Collection) Remove(selector interface{}) error {
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(selector)

	tracing.Trace(c.ctx, database, collection, "remove", args, func() error {
		err = c.collection.Remove(selector)
		return err
	})

	return err
}

// RemoveAll calls mongo collection RemoveAll
func (c *Collection) RemoveAll(selector interface{}) (*mgo.ChangeInfo, error) {
	var result *mgo.ChangeInfo
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(selector)

	tracing.Trace(c.ctx, database, collection, "remove", args, func() error {
		result, err = c.collection.RemoveAll(selector)
		return err
	})

	return result, err
}

// Bulk returns a mongo bulk
func (c *Collection) Bulk() interfaces.Bulk {
	return &Bulk{
		ctx:        c.ctx,
		bulk:       c.collection.Bulk(),
		collection: c.collection,
	}
}

// Bulk holds a monog bulk and implements Bulk interface
type Bulk struct {
	ctx        context.Context
	bulk       *mgo.Bulk
	collection *mgo.Collection
}

// Upsert calls bulk upsert
func (b *Bulk) Upsert(pairs ...interface{}) {
	b.bulk.Upsert(pairs...)
}

// Run executes a bulk run
func (b *Bulk) Run() (*mgo.BulkResult, error) {
	var result *mgo.BulkResult
	var err error

	database := b.collection.Database.Name
	collection := b.collection.FullName
	args := ""

	tracing.Trace(b.ctx, database, collection, "bulkRun", args, func() error {
		result, err = b.bulk.Run()
		return err
	})

	return result, err
}

// Query holds a mongo query and implements Query interface
type Query struct {
	ctx   context.Context
	query *mgo.Query

	database string
	prefix   string
	args     string
}

// Limit calls query Limit
func (q *Query) Limit(n int) interfaces.Query {
	return &Query{
		ctx:      q.ctx,
		query:    q.query.Limit(n),
		database: q.database,
		prefix:   q.prefix,
		args:     q.args,
	}
}

// Iter calls query Iter
func (q *Query) Iter() interfaces.Iter {
	return &Iter{
		iter: q.query.Iter(),
	}
}

// All calls mongo query All
func (q *Query) All(result interface{}) error {
	var err error

	tracing.Trace(q.ctx, q.database, q.prefix, "find", q.args, func() error {
		err = q.query.All(result)
		return err
	})

	return err
}

// One calls mongo query One
func (q *Query) One(result interface{}) error {
	var err error

	tracing.Trace(q.ctx, q.database, q.prefix, "findOne", q.args, func() error {
		err = q.query.One(result)
		return err
	})

	return err
}

// Pipe holds a mongo pipe and implements Pipe interface
type Pipe struct {
	ctx  context.Context
	pipe *mgo.Pipe

	database string
	prefix   string
	args     string
}

// All calls mongo pipe All
func (p *Pipe) All(result interface{}) error {
	var err error

	tracing.Trace(p.ctx, p.database, p.prefix, "aggregate", p.args, func() error {
		err = p.pipe.All(result)
		return err
	})

	return err
}

// Batch calls mongo pipe Batch
func (p *Pipe) Batch(n int) interfaces.Pipe {
	return &Pipe{
		ctx:      p.ctx,
		pipe:     p.pipe.Batch(n),
		database: p.database,
		prefix:   p.prefix,
		args:     p.args,
	}
}

// Iter wraps mongo Iter
type Iter struct {
	iter *mgo.Iter
}

// Next calls mongo iter next
func (i *Iter) Next(result interface{}) bool {
	return i.iter.Next(result)
}

// Close calls mongo iter close
func (i *Iter) Close() error {
	return i.iter.Close()
}

func formatArgs(args ...interface{}) string {
	var array []string

	for _, arg := range args {
		result := fmt.Sprintf("%+v", arg)
		array = append(array, result)
	}

	return strings.Join(array, ", ")
}