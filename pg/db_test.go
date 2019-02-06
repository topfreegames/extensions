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

package pg

import (
	"bytes"
	"context"
	"errors"
	"strings"

	pg "github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/extensions/pg/mocks"
)

var _ = Describe("PG Extension - DB", func() {
	var mockCtrl *gomock.Controller
	var mockDb *mocks.MockDB
	var mockTx *mocks.MockTx

	Describe("[Unit]", func() {
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			mockDb = mocks.NewMockDB(mockCtrl)
			mockTx = mocks.NewMockTx(mockCtrl)
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		Describe("Select", func() {
			It("Should call inner db select", func() {
				db := &DB{inner: mockDb}
				expected := "expected"
				mockDb.EXPECT().Select(expected)
				err := db.Select(expected)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("Insert", func() {
			It("Should call inner db insert", func() {
				db := &DB{inner: mockDb}
				expected1, expected2 := "expected1", "expected2"
				mockDb.EXPECT().Insert(expected1, expected2)
				err := db.Insert(expected1, expected2)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Describe("Update", func() {
			It("Should call inner db update", func() {
				db := &DB{inner: mockDb}
				expected := "expected"
				mockDb.EXPECT().Update(expected)
				err := db.Update(expected)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("Delete", func() {
			It("Should call inner db delete", func() {
				db := &DB{inner: mockDb}
				expected := "expected"
				mockDb.EXPECT().Delete(expected)
				err := db.Delete(expected)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("Model", func() {
			It("Should call inner db model", func() {
				db := &DB{inner: mockDb}
				expected := "expected"
				mockDb.EXPECT().Model(expected).Return(&orm.Query{})
				q := db.Model(expected)
				Expect(q).NotTo(BeNil())
			})
		})

		Describe("Close", func() {
			It("Should call inner db close", func() {
				db := &DB{inner: mockDb}
				mockDb.EXPECT().Close()
				err := db.Close()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("WithContext", func() {
			It("Should call inner db withcontext", func() {
				db := &DB{inner: mockDb}
				expected := &pg.DB{}
				ctx := context.Background()
				mockDb.EXPECT().WithContext(ctx).Return(expected)
				res := db.WithContext(ctx)
				Expect(res).To(Equal(expected))
			})
		})

		Describe("Context", func() {
			It("Should call inner db close", func() {
				db := &DB{inner: mockDb}
				ctx := context.Background()
				mockDb.EXPECT().Context().Return(ctx)
				res := db.Context()
				Expect(res).To(Equal(ctx))
			})
		})

		Describe("CopyFrom", func() {
			It("Should call inner db copy from", func() {
				db := &DB{inner: mockDb}
				expected := "expected"
				reader := strings.NewReader(expected)
				mockDb.EXPECT().CopyFrom(gomock.Any(), expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.CopyFrom(reader, expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})
		})

		Describe("CopyTo", func() {
			It("Should call inner db copy to", func() {
				db := &DB{inner: mockDb}
				expected := "expected"
				var writer bytes.Buffer

				mockDb.EXPECT().CopyTo(gomock.Any(), expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.CopyTo(&writer, expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})
		})

		Describe("FormatQuery", func() {
			It("Should call inner db format query", func() {
				db := &DB{inner: mockDb}
				expected := "expected"

				mockDb.EXPECT().FormatQuery([]byte(expected), expected, expected, expected).Return([]byte(expected))
				res := db.FormatQuery([]byte(expected), expected, expected, expected)
				Expect(res).To(Equal([]byte(expected)))
			})
		})

		Describe("Exec", func() {
			It("Should call inner db exec", func() {
				db := &DB{inner: mockDb}
				expected := "expected"

				mockDb.EXPECT().Context()
				mockDb.EXPECT().Exec(expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.Exec(expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})

			It("Should call inner tx exec", func() {
				db := &DB{inner: mockDb, tx: mockTx}
				expected := "expected"

				mockDb.EXPECT().Context()
				mockTx.EXPECT().Exec(expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.Exec(expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})
		})

		Describe("ExecOne", func() {
			It("Should call inner db exec one", func() {
				db := &DB{inner: mockDb}
				expected := "expected"

				mockDb.EXPECT().Context()
				mockDb.EXPECT().ExecOne(expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.ExecOne(expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})

			It("Should call inner tx exec one", func() {
				db := &DB{inner: mockDb, tx: mockTx}
				expected := "expected"

				mockDb.EXPECT().Context()
				mockTx.EXPECT().ExecOne(expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.ExecOne(expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})
		})

		Describe("Query", func() {
			It("Should call inner db query", func() {
				db := &DB{inner: mockDb}
				expected := "expected"

				mockDb.EXPECT().Context()
				mockDb.EXPECT().Query(expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.Query(expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})

			It("Should call inner tx query", func() {
				db := &DB{inner: mockDb, tx: mockTx}
				expected := "expected"

				mockDb.EXPECT().Context()
				mockTx.EXPECT().Query(expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.Query(expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})
		})

		Describe("QueryOne", func() {
			It("Should call inner db query one", func() {
				db := &DB{inner: mockDb}
				expected := "expected"

				mockDb.EXPECT().Context()
				mockDb.EXPECT().QueryOne(expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.QueryOne(expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})

			It("Should call inner tx query one", func() {
				db := &DB{inner: mockDb, tx: mockTx}
				expected := "expected"

				mockDb.EXPECT().Context()
				mockTx.EXPECT().QueryOne(expected, expected, expected).Return(NewTestResult(nil, 1), nil)
				res, err := db.QueryOne(expected, expected, expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.RowsReturned()).To(Equal(1))
			})
		})

		Describe("Begin", func() {
			It("Should call inner db begin", func() {
				db := &DB{inner: mockDb}

				expected := &pg.Tx{}
				mockDb.EXPECT().Context()
				mockDb.EXPECT().Begin().Return(expected, nil)
				res, err := db.Begin()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(expected))
			})
		})

		Describe("Rollback", func() {
			It("Should call inner tx rollback", func() {
				db := &DB{inner: mockDb, tx: mockTx}

				mockDb.EXPECT().Context()
				mockTx.EXPECT().Rollback()

				err := db.Rollback()
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should fail if no inner tx", func() {
				db := &DB{inner: mockDb}

				expectedError := errors.New("cannot rollback if no transaction")
				mockDb.EXPECT().Context()

				err := db.Rollback()
				Expect(err).To(Equal(expectedError))
			})
		})

		Describe("Commit", func() {
			It("Should call inner tx commit", func() {
				db := &DB{inner: mockDb, tx: mockTx}

				mockDb.EXPECT().Context()
				mockTx.EXPECT().Commit()

				err := db.Commit()
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should fail if no inner tx", func() {
				db := &DB{inner: mockDb}

				expectedError := errors.New("cannot commit if no transaction")
				mockDb.EXPECT().Context()

				err := db.Commit()
				Expect(err).To(Equal(expectedError))
			})
		})

		Describe("DB WithContext", func() {
			It("Should call inner db withcontext and return DB", func() {
				expected := &pg.DB{}
				ctx := context.Background()
				mockDb.EXPECT().WithContext(ctx).Return(expected)
				res := WithContext(ctx, mockDb)
				Expect(res).To(Equal(&DB{inner: expected}))
			})
		})

		Describe("DB Begin", func() {
			It("Should call db begin", func() {
				db := &DB{inner: mockDb}
				expectedTx := &pg.Tx{}
				mockDb.EXPECT().Context()
				mockDb.EXPECT().Begin().Return(expectedTx, nil)

				res, err := Begin(db)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.(*DB).tx).To(Equal(expectedTx))
			})
		})

		Describe("DB Rollback", func() {
			It("Should call db rollback", func() {
				db := &DB{inner: mockDb, tx: mockTx}

				mockDb.EXPECT().Context()
				mockTx.EXPECT().Rollback()

				err := Rollback(db)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should fail if does not implement rollback", func() {
				db := &pg.DB{}

				expectedError := errors.New("db does not implement rollback")
				err := Rollback(db)
				Expect(err).To(Equal(expectedError))
			})
		})

		Describe("DB Commit", func() {
			It("Should call db commit", func() {
				db := &DB{inner: mockDb, tx: mockTx}

				mockDb.EXPECT().Context()
				mockTx.EXPECT().Commit()

				err := Commit(db)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should fail if does not implement commit", func() {
				db := &pg.DB{}

				expectedError := errors.New("db does not implement commit")
				err := Commit(db)
				Expect(err).To(Equal(expectedError))
			})
		})
	})
})
