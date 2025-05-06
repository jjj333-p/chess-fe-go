package gameModes

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"io"
	"net/http"
)

type DbTournament struct {
	TID         int    `json:"tid"`
	Name        string `json:"name"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	MinElo      int    `json:"min_elo"`
	MaxElo      int    `json:"max_elo"`
	Status      string `json:"status"`
	Bracket     int    `json:"bracket"`
	BracketDate string `json:"bracket_date"`
	CanRegister bool   `json:"can_register"`
}

func GetTournaments(token string, serverUrl string) ([]DbTournament, error) {
	// Create the request
	url := fmt.Sprintf("%s/_tournament/list", serverUrl)
	req, err := http.NewRequest("GET", url, nil)
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

	// Parse the response
	var tournaments []DbTournament
	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return tournaments, nil
}

func RegisterForTournament(tournamentID int, authToken, serverUrl string) error {
	// Create the URL for the tournament registration endpoint
	registrationUrl := fmt.Sprintf("%s/_tournament/register/%d", serverUrl, tournamentID)

	// Create a new request
	req, err := http.NewRequest("GET", registrationUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Add the authentication token to the header
	req.Header.Add("token", authToken)

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("authentication required")
		case http.StatusBadRequest:
			return fmt.Errorf("cannot register for tournament: %s", string(body))
		default:
			return fmt.Errorf("server error: %s", string(body))
		}
	}

	return nil
}

func Tournaments(account AccountData, serverUrl string) {
	tournamentsApp := app.New()
	tournamentsWindow := tournamentsApp.NewWindow("Tournaments")

	tournaments, err := GetTournaments(account.AuthToken, serverUrl)
	if err != nil {
		dialog.ShowError(err, tournamentsWindow)
		fmt.Printf("error getting tournaments: %v\n", err)
		return
	}

	// Create slice for grid elements
	gridELS := make([]fyne.CanvasObject, 0)

	// Add headers
	gridELS = append(gridELS,
		widget.NewLabel("ID"),
		widget.NewLabel("Name"),
		widget.NewLabel("Date Range"),
		widget.NewLabel("ELO Range"),
		widget.NewLabel("Status"),
		widget.NewLabel("Action"),
	)

	// Add entries for each tournament
	for _, tournament := range tournaments {
		regBtn := widget.NewButton("Register", func() {
			err := RegisterForTournament(tournament.TID, account.AuthToken, serverUrl)
			if err != nil {
				dialog.ShowError(err, tournamentsWindow)
				return
			}
			tournamentsWindow.Close()
		})
		if !tournament.CanRegister {
			regBtn.Disable()
		}
		gridELS = append(gridELS,
			widget.NewLabel(fmt.Sprintf("%d", tournament.TID)),
			widget.NewLabel(tournament.Name),
			widget.NewLabel(fmt.Sprintf("Start: %s\nEnd: %s",
				tournament.StartDate,
				tournament.EndDate)),
			widget.NewLabel(fmt.Sprintf("%d - %d",
				tournament.MinElo,
				tournament.MaxElo)),
			widget.NewLabel(tournament.Status),
			container.NewVBox(
				regBtn,
				layout.NewSpacer(),
			),
		)
	}

	// Create grid with all elements
	g := container.NewGridWithColumns(6, gridELS...)

	tournamentsWindow.SetContent(container.NewVScroll(g))
	tournamentsWindow.Resize(fyne.NewSize(800, 400))
	tournamentsWindow.ShowAndRun()
}
