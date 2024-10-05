// File: ./internal/bot/giftcode.go

package bot

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Update the baseURL to use the config
var baseURL string

func init() {
	baseURL = config.GiftCode.APIEndpoint
}

func (b *Bot) appendSign(data map[string]string) map[string]string {

	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var str string
	for _, k := range keys {
		str += k + "=" + data[k] + "&"
	}
	str = strings.TrimSuffix(str, "&")

	hash := md5.Sum([]byte(str + b.Config.GiftCode.Salt))
	data["sign"] = hex.EncodeToString(hash[:])
	return data
}

func (b *Bot) RedeemGiftCode(playerID, giftCode string) (bool, string, error) {
	data := map[string]string{
		"fid":  playerID,
		"cdk":  giftCode,
		"time": fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)),
	}

	signedData := b.appendSign(data)

	resp, err := b.makeAPIRequest("/gift_code", signedData)
	if err != nil {
		return false, "", fmt.Errorf("API request failed: %w", err)
	}

	errCode, ok := resp["err_code"].(float64)
	if !ok {
		return false, "", fmt.Errorf("invalid error code format")
	}

	switch int(errCode) {
	case 20000:
		return true, "Gift code redeemed successfully", nil
	case 40014:
		return false, "Gift Code not found", nil
	case 40007:
		return false, "Expired, unable to claim", nil
	case 40008:
		return false, "Gift code already claimed", nil
	default:
		return false, fmt.Sprintf("Unknown error: %v", resp["msg"]), nil
	}
}

func (b *Bot) ValidateGiftCode(giftCode, playerID string) (bool, string) {
	data := map[string]string{
		"fid":  playerID,
		"time": fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)),
	}

	signedData := b.appendSign(data)

	resp, err := b.makeAPIRequest("/player", signedData)
	if err != nil {
		b.logger.WithError(err).Error("Failed to validate player")
		return false, fmt.Sprintf("Failed to validate player: %v", err)
	}

	errCode, ok := resp["err_code"].(float64)
	if !ok {
		b.logger.Error("Invalid error code format")
		return false, "Invalid error code format"
	}

	if int(errCode) != 20000 {
		b.logger.Error("Invalid player ID")
		return false, "Invalid player ID"
	}

	success, message, _ := b.RedeemGiftCode(playerID, giftCode)
	return success, message
}

func (b *Bot) validateGiftCodeLength(code string) bool {
	length := len(code)
	return length >= b.Config.GiftCode.MinLength && length <= b.Config.GiftCode.MaxLength
}

func (b *Bot) makeAPIRequest(endpoint string, data map[string]string) (map[string]interface{}, error) {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}

	req, err := http.NewRequest("POST", baseURL+endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: b.Config.GiftCode.APITimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return result, nil
}
