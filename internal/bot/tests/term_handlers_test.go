// File: internal/bot/tests/term_handlers_test.go

package tests

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestAddTerm(t *testing.T) {
	b, mock := setupMockBot()
	mock.ExpectExec("INSERT INTO terms").WithArgs("term", "description").WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := b.AddTerm(ctx, "term", "description")
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
