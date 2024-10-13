// File: internal/bot/tests/giftcode_test.go

package tests

import (
	"context"
	"testing"

	"the-keeper/internal/bot"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestRedeemGiftCode(t *testing.T) {
	b, mock := setupMockBot()
	mock.ExpectExec("INSERT INTO giftcodes").WithArgs("NEWCODE", "userID", "description").WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	giftCode := bot.GiftCode{
		Code:        "NEWCODE",
		Description: "description",
		Source:      "source",
	}

	err := b.RedeemGiftCode(ctx, "userID", giftCode)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
