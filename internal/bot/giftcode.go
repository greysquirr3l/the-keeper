// File: internal/bot/giftcode.go

package bot

import (
	"context"
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

var baseURL string

// SetGiftCodeBaseURL sets the base URL for gift code API requests.
func SetGiftCodeBaseURL(config *Config) {
	baseURL = config.GiftCode.APIEndpoint
}

// ValidateGiftCode checks if a gift code is valid for a player.
func (b *Bot) ValidateGiftCode(giftCode, playerID string) (bool, string) {
	data := map[string]string{
		"fid":  playerID,
		"cdk":  giftCode,
		"time": fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)),
	}

	signedData := b.appendSign(data)

	resp, err := b.makeAPIRequest(context.Background(), "/gift_code", signedData)
	if err != nil {
		return false, fmt.Sprintf("API request failed: %v", err)
	}

	errCode, ok := resp["err_code"].(float64)
	if !ok {
		return false, "Invalid error code format"
	}

	switch int(errCode) {
	case 20000:
		return true, "Gift code is valid"
	case 40014:
		return false, "Gift Code not found"
	case 40007:
		return false, "Expired, unable to claim"
	case 40008:
		return false, "Gift code already claimed"
	default:
		return false, fmt.Sprintf("Unknown error: %v", resp["msg"])
	}
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

	// Logging the data that will be hashed for signing
	b.logger.Debugf("String to be signed: %s", str)

	hash := md5.Sum([]byte(str + b.Config.GiftCode.Salt))
	signature := hex.EncodeToString(hash[:])

	b.logger.Debugf("Generated signature: %s", signature)

	data["sign"] = signature
	return data
}

// loginPlayer logs in a player using the API.
func (b *Bot) loginPlayer(ctx context.Context, playerID string) error {
	data := map[string]string{
		"fid":  playerID,
		"time": fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)),
	}

	signedData := b.appendSign(data)

	resp, err := b.makeAPIRequest(ctx, "/player", signedData)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}

	if resp["msg"] != "success" {
		return fmt.Errorf("login not possible, validate their player ID")
	}

	b.logger.Debugf("Player %s logged in successfully", playerID)
	return nil
}

// RedeemGiftCode handles the API request to redeem a gift code.
func (b *Bot) RedeemGiftCode(playerID, giftCode string) (bool, string, error) {
	ctx := context.Background()

	// Step 1: Log in the player
	err := b.loginPlayer(ctx, playerID)
	if err != nil {
		return false, "", fmt.Errorf("login failed for player %s: %w", playerID, err)
	}

	// Step 2: Redeem the gift code
	b.logger.Infof("Attempting to redeem gift code '%s' for player ID: %s", giftCode, playerID)
	data := map[string]string{
		"fid":  playerID,
		"cdk":  giftCode,
		"time": fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)),
	}

	signedData := b.appendSign(data)

	resp, err := b.makeAPIRequest(ctx, "/gift_code", signedData)
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

// Make an API request with the given endpoint and data.
func (b *Bot) makeAPIRequest(ctx context.Context, endpoint string, data map[string]string) (map[string]interface{}, error) {
	if b.logger == nil {
		return nil, fmt.Errorf("logger is not initialized")
	}
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}

	fullURL := baseURL + endpoint
	b.logger.Infof("Making request to API endpoint: %s", fullURL)
	b.logger.Debugf("Request data: %v", data)

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: b.Config.GiftCode.APITimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	b.logger.Debugf("API response for endpoint %s: %v", endpoint, result)

	return result, nil
}

// RecordGiftCodeRedemption records a gift code redemption in the database.
func (b *Bot) RecordGiftCodeRedemption(discordID, playerID, giftCode, status string) error {
	redemption := GiftCodeRedemption{
		DiscordID:  discordID,
		PlayerID:   playerID,
		GiftCode:   giftCode,
		Status:     status,
		RedeemedAt: time.Now(),
	}
	result := b.DB.Create(&redemption)
	return result.Error
}

// GetAllGiftCodeRedemptionsPaginated gets all gift code redemptions for admins.
func (b *Bot) GetAllGiftCodeRedemptionsPaginated(page, itemsPerPage int) ([]GiftCodeRedemption, error) {
	var redemptions []GiftCodeRedemption
	offset := (page - 1) * itemsPerPage
	result := b.DB.Order("redeemed_at desc").Offset(offset).Limit(itemsPerPage).Find(&redemptions)
	return redemptions, result.Error
}

// GetUserGiftCodeRedemptionsPaginated gets redemptions for a specific user.
func (b *Bot) GetUserGiftCodeRedemptionsPaginated(discordID string, page, itemsPerPage int) ([]GiftCodeRedemption, error) {
	var redemptions []GiftCodeRedemption
	offset := (page - 1) * itemsPerPage
	result := b.DB.Where("discord_id = ?", discordID).Order("redeemed_at desc").Offset(offset).Limit(itemsPerPage).Find(&redemptions)
	return redemptions, result.Error
}
