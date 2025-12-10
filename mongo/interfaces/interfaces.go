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

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoDB represents the contract for a Mongo DB
type MongoDB interface {
	RunCommand(ctx context.Context, cmd interface{}) *mongo.SingleResult
	Collection(name string) Collection
	Close(ctx context.Context) error
	WithContext(ctx context.Context) MongoDB
}

// Collection represents a mongoDB collection
type Collection interface {
	Find(ctx context.Context, filter interface{}, opts ...any) (Cursor, error)
	FindOne(ctx context.Context, filter interface{}, opts ...any) *mongo.SingleResult
	Aggregate(ctx context.Context, pipeline interface{}, opts ...any) (Cursor, error)
	InsertOne(ctx context.Context, document interface{}, opts ...any) (*mongo.InsertOneResult, error)
	InsertMany(ctx context.Context, documents []interface{}, opts ...any) (*mongo.InsertManyResult, error)
	UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...any) (*mongo.UpdateResult, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...any) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}, opts ...any) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter interface{}, opts ...any) (*mongo.DeleteResult, error)
	BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...any) (*mongo.BulkWriteResult, error)
}

// Cursor wraps mongo Cursor
type Cursor interface {
	Next(ctx context.Context) bool
	Decode(val interface{}) error
	All(ctx context.Context, results interface{}) error
	Close(ctx context.Context) error
}
