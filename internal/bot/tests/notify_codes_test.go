// File: internal/bot/tests/notify_codes_test.go

package tests

import (
	"context"
	"testing"

	"the-keeper/internal/bot"

	"github.com/stretchr/testify/assert"
)

func TestNotifyNewCodes(t *testing.T) {
	b, _ := setupMockBot()
	newCodes := []bot.GiftCode{{
		Code:        "NEWCODE",
		Description: "A new gift code",
		Source:      "MockSite",
	}}

	err := b.notifyNewCodes(context.Background(), newCodes)
	assert.NoError(t, err)
}
