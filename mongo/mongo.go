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

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	tracing "github.com/topfreegames/extensions/v9/tracing/mongo"
)

// Mongo holds the mongo database and connection
// Mongo implements MongoDB interface
type Mongo struct {
	ctx    context.Context
	client *mongo.Client
	db     *mongo.Database
}

// NewMongo return Mongo instance with completed fields
func NewMongo(client *mongo.Client, db *mongo.Database) *Mongo {
	return &Mongo{
		ctx:    context.Background(),
		client: client,
		db:     db,
	}
}

// WithContext creates a shallow copy that uses the given context
func (m *Mongo) WithContext(ctx context.Context) interfaces.MongoDB {
	return &Mongo{
		ctx:    ctx,
		client: m.client,
		db:     m.db,
	}
}

// RunCommand executes a command on the database
func (m *Mongo) RunCommand(ctx context.Context, cmd interface{}) *mongo.SingleResult {
	if ctx == nil {
		ctx = m.ctx
	}

	database := m.db.Name()
	args := formatArgs(cmd)

	var result *mongo.SingleResult
	tracing.Trace(ctx, database, database, "runCommand", args, func() error {
		result = m.db.RunCommand(ctx, cmd)
		return result.Err()
	})

	return result
}

// Collection returns a wrapped collection
func (m *Mongo) Collection(name string) interfaces.Collection {
	return &Collection{
		ctx:        m.ctx,
		collection: m.db.Collection(name),
	}
}

// Close closes mongo client connection
func (m *Mongo) Close(ctx context.Context) error {
	if ctx == nil {
		ctx = m.ctx
	}
	return m.client.Disconnect(ctx)
}

// Collection holds a mongo collection and implements Collection interface
type Collection struct {
	ctx        context.Context
	collection *mongo.Collection
}

// Find executes a find query on Mongo
func (c *Collection) Find(ctx context.Context, filter interface{}, opts ...any) (interfaces.Cursor, error) {
	if ctx == nil {
		ctx = c.ctx
	}

	var cursor *mongo.Cursor
	var err error

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := formatArgs(filter)

	tracing.Trace(ctx, database, collectionName, "find", args, func() error {
		cursor, err = c.collection.Find(ctx, filter)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &Cursor{cursor: cursor}, nil
}

// FindOne executes a findOne query on Mongo
func (c *Collection) FindOne(ctx context.Context, filter interface{}, opts ...any) *mongo.SingleResult {
	if ctx == nil {
		ctx = c.ctx
	}

	var result *mongo.SingleResult

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := formatArgs(filter)

	tracing.Trace(ctx, database, collectionName, "findOne", args, func() error {
		result = c.collection.FindOne(ctx, filter)
		return result.Err()
	})

	return result
}

// Aggregate executes an aggregation pipeline
func (c *Collection) Aggregate(ctx context.Context, pipeline interface{}, opts ...any) (interfaces.Cursor, error) {
	if ctx == nil {
		ctx = c.ctx
	}

	var cursor *mongo.Cursor
	var err error

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := formatArgs(pipeline)

	tracing.Trace(ctx, database, collectionName, "aggregate", args, func() error {
		cursor, err = c.collection.Aggregate(ctx, pipeline)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &Cursor{cursor: cursor}, nil
}

// InsertOne inserts a single document
func (c *Collection) InsertOne(ctx context.Context, document interface{}, opts ...any) (*mongo.InsertOneResult, error) {
	if ctx == nil {
		ctx = c.ctx
	}

	var result *mongo.InsertOneResult
	var err error

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := formatArgs(document)

	tracing.Trace(ctx, database, collectionName, "insertOne", args, func() error {
		result, err = c.collection.InsertOne(ctx, document)
		return err
	})

	return result, err
}

// InsertMany inserts multiple documents
func (c *Collection) InsertMany(ctx context.Context, documents []interface{}, opts ...any) (*mongo.InsertManyResult, error) {
	if ctx == nil {
		ctx = c.ctx
	}

	var result *mongo.InsertManyResult
	var err error

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := formatArgs(documents)

	tracing.Trace(ctx, database, collectionName, "insertMany", args, func() error {
		result, err = c.collection.InsertMany(ctx, documents)
		return err
	})

	return result, err
}

// UpdateByID updates a document by ID
func (c *Collection) UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...any) (*mongo.UpdateResult, error) {
	if ctx == nil {
		ctx = c.ctx
	}

	var result *mongo.UpdateResult
	var err error

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := formatArgs(bson.D{{Key: "_id", Value: id}}, update)

	tracing.Trace(ctx, database, collectionName, "updateOne", args, func() error {
		result, err = c.collection.UpdateByID(ctx, id, update)
		return err
	})

	return result, err
}

// UpdateOne updates a single document
func (c *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...any) (*mongo.UpdateResult, error) {
	if ctx == nil {
		ctx = c.ctx
	}

	var result *mongo.UpdateResult
	var err error

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := formatArgs(filter, update)

	tracing.Trace(ctx, database, collectionName, "updateOne", args, func() error {
		result, err = c.collection.UpdateOne(ctx, filter, update)
		return err
	})

	return result, err
}

// DeleteOne deletes a single document by ID
func (c *Collection) DeleteOne(ctx context.Context, filter interface{}, opts ...any) (*mongo.DeleteResult, error) {
	if ctx == nil {
		ctx = c.ctx
	}

	var result *mongo.DeleteResult
	var err error

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := formatArgs(filter)

	tracing.Trace(ctx, database, collectionName, "deleteOne", args, func() error {
		result, err = c.collection.DeleteOne(ctx, filter)
		return err
	})

	return result, err
}

// DeleteMany deletes multiple documents
func (c *Collection) DeleteMany(ctx context.Context, filter interface{}, opts ...any) (*mongo.DeleteResult, error) {
	if ctx == nil {
		ctx = c.ctx
	}

	var result *mongo.DeleteResult
	var err error

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := formatArgs(filter)

	tracing.Trace(ctx, database, collectionName, "deleteMany", args, func() error {
		result, err = c.collection.DeleteMany(ctx, filter)
		return err
	})

	return result, err
}

// BulkWrite executes multiple write operations
func (c *Collection) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...any) (*mongo.BulkWriteResult, error) {
	if ctx == nil {
		ctx = c.ctx
	}

	var result *mongo.BulkWriteResult
	var err error

	database := c.collection.Database().Name()
	collectionName := c.collection.Name()
	args := ""

	tracing.Trace(ctx, database, collectionName, "bulkWrite", args, func() error {
		result, err = c.collection.BulkWrite(ctx, models)
		return err
	})

	return result, err
}

// Cursor wraps mongo Cursor
type Cursor struct {
	cursor *mongo.Cursor
}

// Next advances the cursor to the next document
func (c *Cursor) Next(ctx context.Context) bool {
	return c.cursor.Next(ctx)
}

// Decode decodes the current document into val
func (c *Cursor) Decode(val interface{}) error {
	return c.cursor.Decode(val)
}

// All retrieves all documents from the cursor
func (c *Cursor) All(ctx context.Context, results interface{}) error {
	return c.cursor.All(ctx, results)
}

// Close closes the cursor
func (c *Cursor) Close(ctx context.Context) error {
	return c.cursor.Close(ctx)
}

func formatArgs(args ...interface{}) string {
	var array []string

	for _, arg := range args {
		result := fmt.Sprintf("%+v", arg)
		array = append(array, result)
	}

	return strings.Join(array, ", ")
}
