package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Global salt variable
const salt = "tB87#kPtkxqOS2"

// Player structure for mapping Discord and Player IDs
type Player struct {
	DiscordID string `yaml:"discord_id"`
	PlayerID  string `yaml:"player_id"`
}

// Main function
func main() {
	// Define flags
	deployFlag := flag.Bool("deploy", false, "Deploy gift codes to a list of players from a YAML file")
	helpFlag := flag.Bool("help", false, "Display available commands")
	flag.Parse()

	if *helpFlag {
		printHelp()
		return
	}

	if *deployFlag {
		if len(flag.Args()) != 2 {
			fmt.Println("Usage: redeem_giftcodes.go --deploy <giftcode> <filename.yml>")
			os.Exit(1)
		}
		giftCode := flag.Arg(0)
		filename := flag.Arg(1)
		err := deployGiftCode(giftCode, filename)
		if err != nil {
			fmt.Printf("Error during deployment: %v\n", err)
			os.Exit(1)
		}
	} else {
		if len(flag.Args()) != 2 {
			fmt.Println("Usage: redeem_giftcodes.go <playerID> <giftcode>")
			os.Exit(1)
		}
		playerID := flag.Arg(0)
		giftCodes := strings.Split(flag.Arg(1), ",") // Support multiple gift codes separated by commas
		for _, giftCode := range giftCodes {
			err := loginAndRedeemGiftCode(playerID, giftCode)
			if err != nil {
				fmt.Printf("Error redeeming gift code %s: %v\n", giftCode, err)
			} else {
				fmt.Printf("Successfully redeemed gift code %s for player %s\n", giftCode, playerID)
			}
		}
	}
}

// Function to print help
func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  redeem_giftcodes.go <playerID> <giftcode>               Redeem a gift code for a player")
	fmt.Println("  redeem_giftcodes.go --deploy <giftcode> <filename.yml>  Deploy a gift code to players listed in a YAML file")
	fmt.Println("  redeem_giftcodes.go --help                              Display available commands")
}

// Function to login a player and then redeem a gift code
func loginAndRedeemGiftCode(playerID, giftCode string) error {
	ctx := context.Background()

	// Step 1: Log in the player
	err := loginPlayer(ctx, playerID)
	if err != nil {
		return fmt.Errorf("login failed for player %s: %v", playerID, err)
	}

	// Step 2: Redeem the gift code
	err = redeemGiftCode(ctx, playerID, giftCode)
	if err != nil {
		return fmt.Errorf("redeem failed for gift code %s: %v", giftCode, err)
	}

	return nil
}

// Login the player to the API
func loginPlayer(ctx context.Context, playerID string) error {
	data := map[string]string{
		"fid":  playerID,
		"time": fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)),
	}

	signedData := appendSign(data)

	resp, err := makeAPIRequest(ctx, "/player", signedData)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}

	if resp["msg"] != "success" {
		return errors.New("login not possible, validate their player ID")
	}

	return nil
}

// RedeemGiftCode handles the API request to redeem a gift code.
func redeemGiftCode(ctx context.Context, playerID, giftCode string) error {
	data := map[string]string{
		"fid":  playerID,
		"cdk":  giftCode,
		"time": fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)),
	}

	signedData := appendSign(data)

	resp, err := makeAPIRequest(ctx, "/gift_code", signedData)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}

	errCode, ok := resp["err_code"].(float64)
	if !ok {
		return errors.New("invalid error code format")
	}

	switch int(errCode) {
	case 20000:
		return nil
	case 40014:
		return errors.New("gift code not found")
	case 40007:
		return errors.New("gift code expired")
	case 40008:
		return errors.New("gift code already claimed")
	default:
		return fmt.Errorf("unknown error: %v", resp["msg"])
	}
}

// Function to deploy gift codes to multiple users from a YAML file
func deployGiftCode(giftCode string, filename string) error {
	var players []Player

	// Load the YAML file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&players)
	if err != nil {
		return fmt.Errorf("failed to decode YAML file: %w", err)
	}

	// Deploy the gift code for each player
	for _, player := range players {
		err := loginAndRedeemGiftCode(player.PlayerID, giftCode)
		if err != nil {
			fmt.Printf("Error for Player ID %s: %v\n", player.PlayerID, err)
		} else {
			fmt.Printf("Successfully redeemed gift code %s for Player ID %s\n", giftCode, player.PlayerID)
		}
	}

	return nil
}

// appendSign appends the necessary signature to the data using the global salt.
func appendSign(data map[string]string) map[string]string {
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

	hash := md5.Sum([]byte(str + salt))
	data["sign"] = hex.EncodeToString(hash[:])
	return data
}

// makeAPIRequest sends a request to the gift code API.
func makeAPIRequest(ctx context.Context, endpoint string, data map[string]string) (map[string]interface{}, error) {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://wos-giftcode-api.centurygame.com/api"+endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 10 * time.Second,
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
