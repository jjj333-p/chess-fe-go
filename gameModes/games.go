package gameModes

import "net/http"
import (
	"encoding/json"
	"fmt"
	"io"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AccountData struct {
	Cred      Credentials
	AuthToken string
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
	var games []string
	if err := json.Unmarshal(body, &games); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return false
	}

	// Print the game IDs
	fmt.Println("Current games:")
	for _, gameID := range games {
		fmt.Println("Game ID:", gameID)
	}

	return resp.StatusCode == http.StatusOK
}
