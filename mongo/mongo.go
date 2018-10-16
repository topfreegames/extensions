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

package mongo

import (
	"context"
	"fmt"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	jaeger "github.com/topfreegames/extensions/jaeger/mongo"
	"github.com/topfreegames/extensions/mongo/interfaces"
)

//Mongo holds the mongo database and connection
// Mongo implements MongoDB interface
type Mongo struct {
	ctx     context.Context
	session *mgo.Session
	db      *mgo.Database
}

//NewMongo return Mongo instance with completed fields
func NewMongo(session *mgo.Session, db *mgo.Database) *Mongo {
	return &Mongo{
		ctx:     nil,
		session: session,
		db:      db,
	}
}

//WithContext creates a shallow copy that uses the given context
func (m *Mongo) WithContext(ctx context.Context) interfaces.MongoDB {
	return &Mongo{
		ctx:     ctx,
		session: m.session,
		db:      m.db,
	}
}

// Run executes run command on database
func (m *Mongo) Run(cmd interface{}, result interface{}) error {
	var err error

	session := m.session.Copy()
	defer session.Close()

	database := m.db.Name
	args := formatArgs(cmd)

	jaeger.Trace(m.ctx, database, database, "runCommand", args, func() error {
		err = m.db.With(session).Run(cmd, result)
		return err
	})

	return err
}

//C returns the collection from databse and a session
// This session needs to be closed afterwards
func (m *Mongo) C(name string) (interfaces.Collection, interfaces.Session) {
	session := m.session.Copy()
	c := &Collection{
		ctx:        m.ctx,
		collection: m.db.With(session).C(name),
	}
	return c, session
}

//Close closes mongo session
func (m *Mongo) Close() {
	m.session.Close()
}

//Collection holds a mongo collection and implements Collection interface
type Collection struct {
	ctx        context.Context
	collection *mgo.Collection
}

//Find executes a find query on Mongo
func (c *Collection) Find(query interface{}) interfaces.Query {
	return &Query{
		ctx:   c.ctx,
		query: c.collection.Find(query),

		database: c.collection.Database.Name,
		prefix:   c.collection.FullName,
		args:     formatArgs(query),
	}
}

//FindId is a conveniene method to execute a find by id query on Mongo
func (c *Collection) FindId(id interface{}) interfaces.Query {
	return &Query{
		ctx:   c.ctx,
		query: c.collection.FindId(id),

		database: c.collection.Database.Name,
		prefix:   c.collection.FullName,
		args:     formatArgs(bson.D{{Name: "_id", Value: id}}),
	}
}

//Insert calls mongo collection Insert
func (c *Collection) Insert(docs ...interface{}) error {
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(docs)

	jaeger.Trace(c.ctx, database, collection, "insert", args, func() error {
		err = c.collection.Insert(docs...)
		return err
	})

	return err
}

//UpsertId calls mongo collection UpsertId
func (c *Collection) UpsertId(id interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	var result *mgo.ChangeInfo
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(bson.D{{Name: "_id", Value: id}}, update)

	jaeger.Trace(c.ctx, database, collection, "updateOne", args, func() error {
		result, err = c.collection.UpsertId(id, update)
		return err
	})

	return result, err
}

//Upsert calls mongo collection Upsert
func (c *Collection) Upsert(selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	var result *mgo.ChangeInfo
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(selector, update)

	jaeger.Trace(c.ctx, database, collection, "updateOne", args, func() error {
		result, err = c.collection.Upsert(selector, update)
		return err
	})

	return result, err
}

//RemoveId calls mongo collection RemoveId
func (c *Collection) RemoveId(id interface{}) error {
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(bson.D{{Name: "_id", Value: id}})

	jaeger.Trace(c.ctx, database, collection, "remove", args, func() error {
		err = c.collection.RemoveId(id)
		return err
	})

	return err
}

//Remove calls mongo collection Remove
func (c *Collection) Remove(selector interface{}) error {
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(selector)

	jaeger.Trace(c.ctx, database, collection, "remove", args, func() error {
		err = c.collection.Remove(selector)
		return err
	})

	return err
}

//RemoveAll calls mongo collection RemoveAll
func (c *Collection) RemoveAll(selector interface{}) (*mgo.ChangeInfo, error) {
	var result *mgo.ChangeInfo
	var err error

	database := c.collection.Database.Name
	collection := c.collection.FullName
	args := formatArgs(selector)

	jaeger.Trace(c.ctx, database, collection, "remove", args, func() error {
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

	jaeger.Trace(b.ctx, database, collection, "bulkRun", args, func() error {
		result, err = b.bulk.Run()
		return err
	})

	return result, err
}

//Query holds a mongo query and implements Query interface
type Query struct {
	ctx   context.Context
	query *mgo.Query

	database string
	prefix   string
	args     string
}

//Iter calls query Iter
func (q *Query) Iter() interfaces.Iter {
	return &Iter{
		iter: q.query.Iter(),
	}
}

//All calls mongo query All
func (q *Query) All(result interface{}) error {
	var err error

	jaeger.Trace(q.ctx, q.database, q.prefix, "find", q.args, func() error {
		err = q.query.All(result)
		return err
	})

	return err
}

//One calls mongo query One
func (q *Query) One(result interface{}) error {
	var err error

	jaeger.Trace(q.ctx, q.database, q.prefix, "findOne", q.args, func() error {
		err = q.query.One(result)
		return err
	})

	return err
}

//Iter wraps mongo Iter
type Iter struct {
	iter *mgo.Iter
}

//Next calls mongo iter next
func (i *Iter) Next(result interface{}) bool {
	return i.iter.Next(result)
}

//Close calls mongo iter close
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
