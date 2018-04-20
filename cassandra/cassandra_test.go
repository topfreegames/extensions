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

package cassandra

import (
	"context"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/cassandra/mocks"
)

type TestQueryObserver struct {
	gocql.ObservedQuery
	DidExecute bool
}

func (obs *TestQueryObserver) ObserveQuery(ctx context.Context, q gocql.ObservedQuery) {
	obs.ObservedQuery = q
	obs.DidExecute = true
}

type TestBatchObserver struct {
	gocql.ObservedBatch
	DidExecute bool
}

func (obs *TestBatchObserver) ObserveBatch(ctx context.Context, b gocql.ObservedBatch) {
	obs.ObservedBatch = b
	obs.DidExecute = true
}

var _ = Describe("Cassandra Extension", func() {
	var config *viper.Viper
	var mockCtrl *gomock.Controller
	var mockDb *mocks.MockDB
	var mockSession *mocks.MockSession

	BeforeEach(func() {
		config = viper.New()
		config.SetConfigFile("../config/test.yaml")
		Expect(config.ReadInConfig()).NotTo(HaveOccurred())
	})

	Describe("[Unit]", func() {
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			mockDb = mocks.NewMockDB(mockCtrl)
			mockSession = mocks.NewMockSession(mockCtrl)
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		Describe("Connect", func() {
			It("Should use config to load connection details", func() {
				params := &ClientParams{
					ClusterConfig: ClusterConfig{
						Prefix: "extensions.cassandra",
					},
					Config:    config,
					CqlOrNil:  mockDb,
					SessOrNil: mockSession,
				}
				client, err := NewClient(params)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Config).NotTo(BeNil())
			})
		})
	})

	Describe("[Integration]", func() {
		Describe("Query with Observer", func() {
			It("Should use config to load connection details", func() {
				obs := &TestQueryObserver{}

				params := &ClientParams{
					ClusterConfig: ClusterConfig{
						Prefix:        "extensions.cassandra",
						QueryObserver: obs,
					},
					Config: config,
				}

				client, err := NewClient(params)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Config).NotTo(BeNil())

				stmt := "SELECT now() FROM system.local"
				err = client.Session.Query(stmt).Exec()
				Expect(err).NotTo(HaveOccurred())

				Expect(obs.DidExecute).To(Equal(true))
				Expect(obs.Keyspace).To(Equal("test"))
				Expect(obs.Statement).To(Equal(stmt))
			})
		})
		Describe("Barch with Observer", func() {
			It("Should use config to load connection details", func() {
				obs := &TestBatchObserver{}

				params := &ClientParams{
					ClusterConfig: ClusterConfig{
						Prefix:        "extensions.cassandra",
						BatchObserver: obs,
					},
					Config: config,
				}

				client, err := NewClient(params)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Config).NotTo(BeNil())

				batch := client.Session.NewBatch(gocql.LoggedBatch)

				stmt1 := "INSERT INTO user (id, info) VALUES ('1', 'User with id 1')"
				stmt2 := "INSERT INTO user (id, info) VALUES ('2', 'User with id 2')"
				batch.Query(stmt1)
				batch.Query(stmt2)

				err = client.Session.ExecuteBatch(batch)
				Expect(err).NotTo(HaveOccurred())

				Expect(obs.DidExecute).To(Equal(true))
				Expect(obs.Keyspace).To(Equal("test"))
				Expect(len(obs.Statements)).To(Equal(2))
				Expect(obs.Statements[0]).To(Equal(stmt1))
				Expect(obs.Statements[1]).To(Equal(stmt2))
			})
		})
	})
})
