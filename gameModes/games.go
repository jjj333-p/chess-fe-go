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
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync/atomic"
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

	// Create Fyne application and window
	a := app.New()
	w := a.NewWindow("Current Games")

	var desiredGameIDToPlay int

	//init slice of length for the elements
	gridELS := make([]fyne.CanvasObject, 0, 6*(len(games)+1))

	gridELS = append(gridELS,
		widget.NewLabel("Game ID"),
		widget.NewLabel("White Player"),
		widget.NewLabel("Black Player"),
		widget.NewLabel("Status"),
		widget.NewLabel("Tournament"),
		widget.NewLabel("Action"),
	)

	// Create entries for each game
	for _, game := range games {
		gridELS = append(gridELS,
			widget.NewLabel(fmt.Sprintf("%d", game.GameID)),
			widget.NewLabel(fmt.Sprintf("%s (%d elo)", game.WhiteName, game.WhiteElo)),
			widget.NewLabel(fmt.Sprintf("%s (%d elo)", game.BlackName, game.BlackElo)),
			widget.NewLabel(game.Status),
			widget.NewLabel(game.TName),
			container.NewVBox(
				widget.NewButton("Play Now", func() {
					desiredGameIDToPlay = game.GameID
					w.Close()
				}),
				layout.NewSpacer(),
			),
		)

	}

	fmt.Println(len(gridELS))

	//create grid with all els
	g := container.NewGridWithColumns(6, gridELS...)

	w.SetContent(container.NewVScroll(g))

	w.Resize(fyne.NewSize(800, 200))
	w.ShowAndRun()

	fmt.Printf("game to play: %d\n", desiredGameIDToPlay)

	// Find the selected game
	var selectedGame *DbGame
	for _, game := range games {
		if game.GameID == desiredGameIDToPlay {
			gameCopy := game
			selectedGame = &gameCopy
			break
		}
	}

	if selectedGame == nil {
		panic("No game found with ID:" + strconv.Itoa(desiredGameIDToPlay))
	}

	fmt.Printf("White: %s, Black: %s\n", selectedGame.WhiteName, selectedGame.BlackName)

	dbmoves := selectedGame.Moves
	sort.Slice(dbmoves, func(i, j int) bool {
		return dbmoves[i].MIndex < dbmoves[j].MIndex
	})

	gameApp := app.New()
	gameWindow := gameApp.NewWindow("Practice Game")

	viewingHistorical := atomic.Bool{}
	viewedMove := atomic.Int32{}

	moves := make([]chessboard.Move, 0)

	board := chessboard.NewChessBoard()

	playingText := widget.NewLabel("White's game...")
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
			board.MovePiece(move.To, move.From)
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
			board.MovePiece(move.From, move.To)
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

	for _, dbmove := range dbmoves {
		//append(moves, chessboard.Move{
		//	From: chessboard.Location{Rank: }
		//})
		from := &chessboard.Location{
			Rank: dbmove.MFrom % 8,
			File: dbmove.MFrom / 8,
		}
		to := &chessboard.Location{
			Rank: dbmove.MTo % 8,
			File: dbmove.MTo / 8,
		}
		moves = append(moves, chessboard.Move{
			From: from,
			To:   to,
		})
		board.MovePiece(from, to)
	}
	viewedMove.Store(int32(len(moves)))
	updateViewingText()
	isBlack := selectedGame.BlackName == account.Cred.Username

	//go func() {
	//
	//}()

	//board.PrepareForMove()

	gameWindow.ShowAndRun()

	return true
}
