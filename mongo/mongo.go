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
	return m.db.With(session).C(name), session
}

//Close closes mongo session
func (m *Mongo) Close() {
	m.session.Close()
}
