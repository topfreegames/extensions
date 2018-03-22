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

package interfaces

import (
	"context"

	mgo "gopkg.in/mgo.v2"
)

//MongoDB represents the contract for a Mongo DB
type MongoDB interface {
	Run(cmd interface{}, result interface{}) error
	C(name string) (Collection, Session)
	Close()
	WithContext(ctx context.Context) MongoDB
}

//Collection represents a mongoDB collection
type Collection interface {
	Find(query interface{}) Query
	FindId(id interface{}) Query
	Insert(docs ...interface{}) error
	UpsertId(id interface{}, update interface{}) (*mgo.ChangeInfo, error)
	RemoveId(id interface{}) error
}

//Session is the mongoDB session
type Session interface {
	Copy() *mgo.Session
	Close()
}

//Query wraps mongo Query
type Query interface {
	Iter() Iter
	All(result interface{}) error
	One(result interface{}) error
}

//Iter wraps mongo Iter
type Iter interface {
	Next(result interface{}) bool
	Close() error
}
