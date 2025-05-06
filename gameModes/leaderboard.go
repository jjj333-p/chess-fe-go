package gameModes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"io"
)

type LeaderboardEntry struct {
	UID      int     `json:"uid"`
	Username string  `json:"username"`
	Rating   float64 `json:"rating"`
	Rank     int     `json:"rank"`
}

func GetCurrentLeaderboard(token string, serverUrl string) ([]LeaderboardEntry, error) {
	// Create the request
	req, err := http.NewRequest("GET", serverUrl+"/_leaderboard/current", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set the authentication token
	req.Header.Set("token", token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication required")
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var leaderboard []LeaderboardEntry
	if err := json.NewDecoder(resp.Body).Decode(&leaderboard); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return leaderboard, nil
}

func Leaderboards(account AccountData, serverUrl string) {
	leaderboardApp := app.New()
	leaderboardWindow := leaderboardApp.NewWindow("Leaderboard")

	leaderboard, err := GetCurrentLeaderboard(account.AuthToken, serverUrl)
	if err != nil {
		dialog.ShowError(err, leaderboardWindow)
		fmt.Printf("error getting leaderboard: %v\n", err)
		return
	}

	// Create slice for grid elements
	gridELS := make([]fyne.CanvasObject, 0)

	// Add headers
	gridELS = append(gridELS,
		widget.NewLabel("Rank"),
		widget.NewLabel("Username"),
		widget.NewLabel("Rating"),
		widget.NewLabel("UID"),
	)

	// Add entries for each player
	for _, entry := range leaderboard {
		gridELS = append(gridELS,
			widget.NewLabel(fmt.Sprintf("%d", entry.Rank)),
			widget.NewLabel(entry.Username),
			widget.NewLabel(fmt.Sprintf("%.2f", entry.Rating)),
			widget.NewLabel(strconv.Itoa(entry.UID)),
		)
	}

	// Create grid with all elements
	g := container.NewGridWithColumns(4, gridELS...)

	leaderboardWindow.SetContent(container.NewVScroll(g))
	leaderboardWindow.Resize(fyne.NewSize(600, 400))
	leaderboardWindow.ShowAndRun()
}
