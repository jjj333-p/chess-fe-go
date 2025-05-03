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
	"github.com/jjj333-p/chess-fe-go/chessboard"
	"net/http"
)

const serverUrl = "http://localhost:9001"

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token  string `json:"token"`
	ErrStr string `json:"error"`
}

type accountData struct {
	cred      credentials
	authToken string
}

func login(creds credentials) loginResponse {
	jsonData, _ := json.Marshal(creds)

	resp, err := http.Post(serverUrl+"/_login/authenticate",
		"application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return loginResponse{ErrStr: err.Error()}
	}
	defer resp.Body.Close()

	var authn loginResponse
	err = json.NewDecoder(resp.Body).Decode(&authn)
	if err != nil {
		return loginResponse{ErrStr: err.Error()}
	}
	return authn
}

func practiceGame() bool {
	gameApp := app.New()
	gameWindow := gameApp.NewWindow("Practice Game")

	board := chessboard.NewChessBoard()

	gameWindow.SetContent(board.Grid)

	gameWindow.Resize(fyne.NewSize(400, 400))

	go func() {

		var startPosChan chan *chessboard.Location

		for ok := true; ok; ok = startPosChan == nil {
			fyne.DoAndWait(func() { startPosChan = board.PrepareForMove(false, true) })
		}
		startPos := <-startPosChan
		fmt.Println(startPos, "startPos")
		var endPosChan chan *chessboard.Location
		for ok := true; ok; ok = endPosChan == nil {
			fyne.DoAndWait(func() { endPosChan = board.MoveChooser(startPos.Rank, startPos.File) })
			fmt.Println(endPosChan)
		}
		endPos := <-endPosChan
		fmt.Println(endPos.Rank, endPos.File)

	}()

	gameWindow.ShowAndRun()
	return true
}

func main() {
	loginApp := app.New()
	loginWindow := loginApp.NewWindow("Login")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loadingSpinner := widget.NewProgressBarInfinite()
	loadingSpinner.Hide()

	var account accountData

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

		creds := credentials{
			Username: usernameEntry.Text,
			Password: passwordEntry.Text,
		}
		account.cred = creds
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

		account.authToken = authn.Token
		loadingSpinner.Hide()
		loginBtn.Enable()
		dialog.ShowInformation("Success", "Login successful!", loginWindow)

		fmt.Println("Login successful!", account.authToken)
		loginWindow.Close()

	})

	practiceGameInsteadOfLogin := false
	localGameBtn := widget.NewButton("Local Game", func() {
		loginWindow.Close()
		practiceGameInsteadOfLogin = true
	})

	loginVBox := container.NewVBox(
		layout.NewSpacer(),
		usernameEntry,
		passwordEntry,
		loginBtn,
		localGameBtn,
		loadingSpinner,
		layout.NewSpacer(),
	)

	loginWindow.SetContent(loginVBox)
	loginWindow.Resize(fyne.NewSize(300, 200))
	loginWindow.ShowAndRun()

	if practiceGameInsteadOfLogin {
		practiceGame()
		return
	}

	//if the user exits dont bring up the next ui
	if account.authToken == "" {
		return
	}

	for {
		menuChoice := 0
		menuApp := app.New()
		menuWindow := menuApp.NewWindow("Menu")

		practiceBtn := widget.NewButton("Practice Game", func() {
			menuChoice = 1
			menuWindow.Close()
		})

		onlineGameBtn := widget.NewButton("New Online Game", func() {
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
			returnToMenu = practiceGame()
		case 2:
			fmt.Println("New Online Game")
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
