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
	"github.com/jjj333-p/chess-fe-go/chessboard"
	"net/http"
	"strconv"
	"sync/atomic"
)

func GetOldGames(token string, serverUrl string) ([]DbGame, error) {
	// Create the request
	req, err := http.NewRequest("GET", serverUrl+"/_game/old", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add the authentication token
	req.Header.Set("token", token)

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Check for authentication error
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication required")
	}

	// Check for other errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error status %d", resp.StatusCode)
	}

	// Parse the response
	var games []DbGame
	if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return games, nil
}

func OldGames(account AccountData, serverUrl string) {
	oldGamesApp := app.New()
	oldGamesWindow := oldGamesApp.NewWindow("Past Games")

	games, err := GetOldGames(account.AuthToken, serverUrl)
	if err != nil {
		fmt.Printf("Error fetching old games: %v\n", err)
		return
	}

	// Create slice for grid elements
	gridELS := make([]fyne.CanvasObject, 0)

	// Add headers
	gridELS = append(gridELS,
		widget.NewLabel("Game ID"),
		widget.NewLabel("Date"),
		widget.NewLabel("White Player"),
		widget.NewLabel("Black Player"),
		widget.NewLabel("Status"),
		widget.NewLabel("Tournament"),
		widget.NewLabel(""),
	)

	viewGID := 0

	// Add entries for each game
	for _, game := range games {
		gID := game.GameID
		gridELS = append(gridELS,
			widget.NewLabel(fmt.Sprintf("%d", game.GameID)),
			widget.NewLabel(game.Date),
			widget.NewLabel(fmt.Sprintf("%s (%d)", game.WhiteName, game.WhiteElo)),
			widget.NewLabel(fmt.Sprintf("%s (%d)", game.BlackName, game.BlackElo)),
			widget.NewLabel(game.Status),
			widget.NewLabel(game.TName),
			widget.NewButton("View History", func() {
				viewGID = gID
				oldGamesWindow.Close()
			}),
		)
	}

	// Create grid with all elements
	g := container.NewGridWithColumns(7, gridELS...)

	oldGamesWindow.SetContent(container.NewVScroll(g))
	oldGamesWindow.Resize(fyne.NewSize(800, 400))
	oldGamesWindow.ShowAndRun()

	if viewGID == 0 {
		return
	}

	var selectedGame DbGame
	for _, game := range games {
		if game.GameID == viewGID {
			selectedGame = game
			break
		}
	}

	fmt.Printf("Selected game: ID=%d, White=%s(%d) vs Black=%s(%d), Status=%s, Tournament=%s, Date=%s\n",
		selectedGame.GameID,
		selectedGame.WhiteName, selectedGame.WhiteElo,
		selectedGame.BlackName, selectedGame.BlackElo,
		selectedGame.Status,
		selectedGame.TName,
		selectedGame.Date)

	gameApp := app.New()
	gameWindow := gameApp.NewWindow("Historical Game")

	viewingHistorical := atomic.Bool{}
	viewedMove := atomic.Int32{}

	moves := make([]chessboard.Move, 0)

	board := chessboard.NewChessBoard()

	var statusString string
	if selectedGame.Status == "W" {
		statusString = "White won the game."
	} else if selectedGame.Status == "B" {
		statusString = "Black won the game."
	} else {
		statusString = "Unknown status \"" + selectedGame.Status + "\""
	}

	playingText := widget.NewLabel(statusString)

	viewingText := widget.NewLabel("Viewing move 0 of 0")
	updateViewingText := func() {
		fyne.Do(func() {
			viewingText.SetText("Viewing move " + strconv.Itoa(int(viewedMove.Load())) + " of " + strconv.Itoa(len(moves)))
		})
	}

	displayPrevState := func() {
		moveNo := viewedMove.Load()
		fmt.Println("moveNo", moveNo)
		if moveNo == 0 {
			message := "You were already viewing the initial state. There is no earlier move to view."
			fyne.Do(func() {
				dialog.ShowInformation("No move", message, gameWindow)
			})
			return
		}

		//dont make new changes onto an old board
		viewingHistorical.Store(true)
		move := moves[moveNo-1]
		viewedMove.Add(-1)

		//undo change to board
		fyne.Do(func() {
			board.MovePiece(move.To, move.From, true)
		})

		updateViewingText()
	}
	displayNextState := func() {
		moveNo := viewedMove.Load()
		fmt.Println("moveNo", moveNo)
		if moveNo == int32(len(moves)) {
			message := "You were already viewing the latest state. There is no newer move to view."
			fyne.Do(func() {
				dialog.ShowInformation("No move", message, gameWindow)
			})
			return
		}
		viewingHistorical.Store(true)
		move := moves[moveNo]
		viewedMove.Add(1)

		//if we are on current move we are no longer in historical, start updating again
		if viewedMove.Load() == int32(len(moves)) {
			viewingHistorical.Store(false)
		}

		//redo change to board
		fyne.Do(func() {
			board.MovePiece(move.From, move.To, false)
		})

		updateViewingText()
	}

	prevButton := widget.NewButton("◀️", displayPrevState)
	nextButton := widget.NewButton("▶️", displayNextState)
	doublePrev := widget.NewButton("⏪️", func() {
		vm := viewedMove.Load()
		if vm == 0 {
			dialog.ShowInformation("No Moves", "You are already viewing the first move.", gameWindow)
		}
		for _ = range vm {
			displayPrevState()
		}
	})
	doubleNext := widget.NewButton("⏩️", func() {
		vm := -1 * (viewedMove.Load() - int32(len(moves)))
		fmt.Println("moves", len(moves), "vm", vm)
		if vm < 1 {
			dialog.ShowInformation("No Moves", "You are already viewing the latest move.", gameWindow)
		}
		for _ = range vm {
			displayNextState()
		}
	})

	topBar := container.NewHBox(playingText, layout.NewSpacer(), doublePrev, prevButton, viewingText, nextButton, doubleNext)

	content := container.NewVBox(topBar, board.Grid)

	gameWindow.SetContent(content)

	gameWindow.Resize(fyne.NewSize(400, 400))

	for _, dbmove := range selectedGame.Moves {
		mv := dbMoveToMove(&dbmove)
		moves = append(moves, mv)
		board.MovePiece(mv.From, mv.To, false)
	}
	viewedMove.Store(int32(len(moves)))
	updateViewingText()
	fmt.Println(moves)
	isBlack := selectedGame.BlackName == account.Cred.Username
	board.PrepareForMove(isBlack, false)
	board.DisableAllBtn()

	gameWindow.ShowAndRun()

}
