package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/jjj333-p/chess-fe-go/gameModes"
	"net/http"
)

const serverUrl = "http://66.91.172.196:5000"

type loginResponse struct {
	Token  string `json:"token"`
	ErrStr string `json:"error"`
}

func login(creds gameModes.Credentials) loginResponse {
	// First get the token (existing code)
	resp, err := http.Get(serverUrl + "/request_token")
	if err != nil {
		fmt.Println("Error requesting token:", err)
		return loginResponse{ErrStr: "Error requesting token"}
	}
	defer resp.Body.Close()

	var tokenResp struct {
		Token string `json:"token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return loginResponse{ErrStr: "Error requesting token"}
	}

	fmt.Println("Token:", tokenResp.Token)

	jsonData, err := json.Marshal(creds)
	if err != nil {
		fmt.Println("Error marshaling credentials:", err)
		return loginResponse{ErrStr: "Error preparing login request"}
	}

	// Create authentication request
	req, err := http.NewRequest("POST", serverUrl+"/_login/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return loginResponse{ErrStr: "Error creating login request"}
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", tokenResp.Token)

	// Send the authentication request
	client := &http.Client{}
	authResp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error during authentication:", err)
		return loginResponse{ErrStr: "Error during authentication"}
	}
	defer authResp.Body.Close()

	// Check response status
	if authResp.StatusCode == 401 {
		return loginResponse{ErrStr: "Invalid credentials"}
	}
	if authResp.StatusCode != 200 {
		return loginResponse{ErrStr: fmt.Sprintf("Server error: %d", authResp.StatusCode)}
	}

	// Since authentication was successful, return the initial token
	return loginResponse{Token: tokenResp.Token}
}

func main() {

	initialChoice := 0
	initialApp := app.New()
	initialWindow := initialApp.NewWindow("Authentication")

	loginchBtn := widget.NewButton("Login", func() {
		initialChoice = 1
		initialWindow.Close()
	})

	registerBtn := widget.NewButton("Register", func() {
		initialChoice = 2
		initialWindow.Close()
	})

	localGameBtn := widget.NewButton("Local Game", func() {
		initialChoice = 3
		initialWindow.Close()
	})

	initialContent := container.NewCenter(container.NewVBox(
		layout.NewSpacer(),
		loginchBtn,
		registerBtn,
		localGameBtn,
		layout.NewSpacer(),
	))

	initialWindow.SetContent(initialContent)
	initialWindow.Resize(fyne.NewSize(200, 100))
	initialWindow.ShowAndRun()

	var account gameModes.AccountData

	// Handle the choice
	switch initialChoice {
	case 1:
		// Handle login
		println("Login selected")

		loginApp := app.New()
		loginWindow := loginApp.NewWindow("Login")

		usernameEntry := widget.NewEntry()
		usernameEntry.SetPlaceHolder("Username")
		passwordEntry := widget.NewPasswordEntry()
		passwordEntry.SetPlaceHolder("Password")

		loadingSpinner := widget.NewProgressBarInfinite()
		loadingSpinner.Hide()

		var loginBtn *widget.Button
		loginBtn = widget.NewButton("Login", func() {
			if usernameEntry.Text == "" {
				dialog.ShowError(errors.New("please enter a username"), loginWindow)
				return
			}
			if passwordEntry.Text == "" {
				dialog.ShowError(errors.New("please enter a password"), loginWindow)
				return
			}

			loadingSpinner.Show()
			loginBtn.Disable()

			creds := gameModes.Credentials{
				Username: usernameEntry.Text,
				Password: passwordEntry.Text,
			}
			account.Cred = creds
			authn := login(creds)

			if authn.ErrStr != "" {
				loadingSpinner.Hide()
				loginBtn.Enable()
				dialog.ShowError(errors.New(authn.ErrStr), loginWindow)
				return
			} else if authn.Token == "" {
				loadingSpinner.Hide()
				loginBtn.Enable()
				dialog.ShowError(errors.New("empty token response from server"), loginWindow)
				return
			}

			account.AuthToken = authn.Token
			loadingSpinner.Hide()
			loginBtn.Enable()
			dialog.ShowInformation("Success", "Login successful!", loginWindow)

			fmt.Println("Login successful!", account.AuthToken)
			loginWindow.Close()

		})

		loginVBox := container.NewVBox(
			layout.NewSpacer(),
			usernameEntry,
			passwordEntry,
			loginBtn,
			loadingSpinner,
			layout.NewSpacer(),
		)

		loginWindow.SetContent(loginVBox)
		loginWindow.Resize(fyne.NewSize(300, 200))
		loginWindow.ShowAndRun()

		//if the user exits dont bring up the next ui
		if account.AuthToken == "" {
			return
		}
	case 2:
		// Handle registration
		println("Register selected")
	case 3:
		gameModes.PracticeGame()
		return
	case 0:
		// Window was closed
		println("Window closed")
		return
	}

	for {
		menuChoice := 0
		menuApp := app.New()
		menuWindow := menuApp.NewWindow("Menu")

		practiceBtn := widget.NewButton("Local Game", func() {
			menuChoice = 1
			menuWindow.Close()
		})

		onlineGameBtn := widget.NewButton("Online Game", func() {
			menuChoice = 2
			menuWindow.Close()
		})

		pastGamesBtn := widget.NewButton("View Past Games", func() {
			menuChoice = 3
			menuWindow.Close()
		})

		menuContent := container.NewCenter(container.NewVBox(
			practiceBtn,
			onlineGameBtn,
			pastGamesBtn,
		))

		menuWindow.SetContent(menuContent)
		menuWindow.ShowAndRun()

		returnToMenu := false
		switch menuChoice {
		case 0:
			return
		case 1:
			fmt.Println("Practice Game")
			returnToMenu = gameModes.PracticeGame()
		case 2:
			fmt.Println("New Online Game")
			returnToMenu = gameModes.Games(account, serverUrl)
		case 3:
			fmt.Println("View Past Games")
		default:
			panic("Unknow menu choice")
		}
		if !returnToMenu {
			return
		}
	}
}
