package pg

import "github.com/go-pg/pg/orm"

// TestResult struct
type TestResult struct {
	rowsAffected int
	rowsReturned int
	err          error
}

// Model test method
func (t *TestResult) Model() orm.Model {
	return nil
}

// RowsReturned test method
func (t *TestResult) RowsReturned() int {
	return t.rowsReturned
}

// RowsAffected test method
func (t *TestResult) RowsAffected() int {
	return t.rowsAffected
}

// NewTestResult ctor
func NewTestResult(err error, vals int) *TestResult {
	return &TestResult{
		rowsAffected: vals,
		rowsReturned: vals,
		err:          err,
	}
}
