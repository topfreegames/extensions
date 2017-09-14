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
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2/bson"

	. "github.com/topfreegames/extensions/mongo"
	"github.com/topfreegames/extensions/mongo/interfaces"

	. "github.com/onsi/ginkgo"
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
		})

		Describe("Close", func() {
			It("Should close after creating", func() {
				client, err := NewClient("extensions.mongo", config, mockDb)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().Close()
				client.Close()
			})
		})

		Describe("Run", func() {
			It("Should execute command with run", func() {
				collectionName := "coll"

				client, err := NewClient("extensions.mongo", config, mockDb)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().Close()
				mockDb.EXPECT().C(collectionName).Return(mockColl, nil)
				mockColl.EXPECT().Insert(gomock.Any())

				c, _ := client.MongoDB.C(collectionName)
				err = c.Insert(bson.M{})
				Expect(err).NotTo(HaveOccurred())
				client.Close()
			})

			It("Should execute Run command", func() {
				client, err := NewClient("extensions.mongo", config, mockDb)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().Run(gomock.Any(), gomock.Any())

				var result string
				err = client.MongoDB.Run(bson.D{
					{"create", "mycollection"},
				}, &result)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
