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
	"github.com/topfreegames/extensions/mongo/interfaces"
	mgo "gopkg.in/mgo.v2"
)

//Mongo holds the mongo database and connection
// Mongo implements MongoDB interface
type Mongo struct {
	session *mgo.Session
	db      *mgo.Database
}

//NewMongo return Mongo instance with completed fields
func NewMongo(session *mgo.Session, db *mgo.Database) *Mongo {
	return &Mongo{
		session: session,
		db:      db,
	}
}

// Run executes run command on database
func (m *Mongo) Run(cmd interface{}, result interface{}) error {
	session := m.session.Copy()
	defer session.Close()
	return m.db.With(session).Run(cmd, result)
}

//C returns the collection from databse and a session
// This session needs to be closed afterwards
func (m *Mongo) C(name string) (interfaces.Collection, interfaces.Session) {
	session := m.session.Copy()
	return &Collection{collection: m.db.With(session).C(name)}, session
}

//Clone clones mongo session
func (m *Mongo) Clone() interfaces.MongoDB {
	if m == nil {
		return nil
	}

	return &Mongo{
		m.session.Clone(),
		m.db,
	}
}

//Close closes mongo session
func (m *Mongo) Close() {
	m.session.Close()
}

//Collection holds a mongo collection and implements Collection interface
type Collection struct {
	collection *mgo.Collection
}

//Find executes a find query on Mongo
func (c *Collection) Find(query interface{}) interfaces.Query {
	return &Query{
		query: c.collection.Find(query),
	}
}

func (c *Collection) FindId(id interface{}) interfaces.Query {
	return &Query{
		query: c.collection.FindId(id),
	}
}

//Insert calls mongo collection Insert
func (c *Collection) Insert(docs ...interface{}) error {
	return c.collection.Insert(docs...)
}

func (c *Collection) UpsertId(id interface{}, update interface{}) (info *mgo.ChangeInfo, err error) {
	return c.collection.UpsertId(id, update)
}

func (c *Collection) RemoveId(id interface{}) error {
	return c.collection.RemoveId(id)
}

//Query holds a mongo query and implements Query interface
type Query struct {
	query *mgo.Query
}

//Iter calls query Iter
func (q *Query) Iter() interfaces.Iter {
	return &Iter{
		iter: q.query.Iter(),
	}
}

//All calls mongo query All
func (q *Query) All(result interface{}) error {
	return q.query.All(result)
}

func (q *Query) One(result interface{}) error {
	return q.query.One(result)
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
