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

package mongo_test

import (
	"context"

	"go.uber.org/mock/gomock"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/bson"

	. "github.com/topfreegames/extensions/v9/mongo"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mongo", func() {
	var config *viper.Viper
	var mockCtrl *gomock.Controller
	var mockDb *interfaces.MockMongoDB
	var mockColl *interfaces.MockCollection

	BeforeEach(func() {
		config = viper.New()
		config.SetConfigFile("../config/test.yaml")
		Expect(config.ReadInConfig()).NotTo(HaveOccurred())
	})

	Describe("[Unit]", func() {
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			mockDb = interfaces.NewMockMongoDB(mockCtrl)
			mockColl = interfaces.NewMockCollection(mockCtrl)
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		Describe("Connect", func() {
			It("Should use config to load connection details", func() {
				_, err := NewClient("extensions.mongo", config, mockDb)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should return error if connection fails", func() {
				config.Set("extensions.mongo.url", "mongodb://invalid-host:27017")
				config.Set("extensions.mongo.connectionTimeout", "100ms")
				client, err := NewClient("extensions.mongo", config)
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
			})
		})

		Describe("Close", func() {
			It("Should close after creating", func() {
				client, err := NewClient("extensions.mongo", config, mockDb)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().Close(gomock.Any()).Return(nil)
				err = client.Close()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("Operations", func() {
			It("Should execute insert with InsertOne", func() {
				collectionName := "coll"
				ctx := context.Background()

				client, err := NewClient("extensions.mongo", config, mockDb)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().Collection(collectionName).Return(mockColl)
				mockColl.EXPECT().InsertOne(ctx, bson.M{"test": "data"}).Return(nil, nil)

				c := client.MongoDB.Collection(collectionName)
				_, err = c.InsertOne(ctx, bson.M{"test": "data"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should execute RunCommand", func() {
				ctx := context.Background()
				client, err := NewClient("extensions.mongo", config, mockDb)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().RunCommand(ctx, bson.D{{Key: "create", Value: "mycollection"}}).Return(nil)

				client.MongoDB.RunCommand(ctx, bson.D{
					{Key: "create", Value: "mycollection"},
				})
			})
		})
	})
})
