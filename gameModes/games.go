package gameModes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AccountData struct {
	Cred      Credentials
	AuthToken string
}

type DbMove struct {
	MIndex    int    `json:"mindex"`
	GameID    int    `json:"game_id"`
	MFrom     int    `json:"mfrom"`
	MTo       int    `json:"mto"`
	PieceName string `json:"piece_name"`
}

type DbGame struct {
	GameID         int      `json:"game_id"`
	Date           string   `json:"date"`
	Status         string   `json:"status"`
	WhiteID        int      `json:"white_id"`
	WhiteName      string   `json:"white_name"`
	BlackID        int      `json:"black_id"`
	BlackName      string   `json:"black_name"`
	WhiteElo       int      `json:"white_elo"`
	BlackElo       int      `json:"black_elo"`
	WhiteEloChange int      `json:"white_elo_change"`
	BlackEloChange int      `json:"black_elo_change"`
	TID            int      `json:"tid"`
	Bracket        int      `json:"bracket"`
	Turn           string   `json:"turn"`
	TName          string   `json:"tname"`
	Moves          []DbMove `json:"moves"`
}

func Games(account AccountData, serverUrl string) bool {
	// Create a new request
	req, err := http.NewRequest("GET", serverUrl+"/_game/current", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return false
	}

	// Add the authentication token to headers
	req.Header.Add("token", account.AuthToken)

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return false
	}

	// Parse the JSON response
	var games []DbGame
	if err := json.Unmarshal(body, &games); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return false
	}

	// Print the games information
	fmt.Println("Current games:")
	for _, game := range games {
		fmt.Printf("Game ID: %d, White: %s vs Black: %s, Status: %s\n",
			game.GameID, game.WhiteName, game.BlackName, game.Status)
	}

	return resp.StatusCode == http.StatusOK
}
