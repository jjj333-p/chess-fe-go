package gameModes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/jjj333-p/chess-fe-go/chessboard"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync/atomic"
	"time"
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

func fetchLastMove(gameID int, token string, serverUrl string) (*DbMove, error) {
	url := fmt.Sprintf("%s/_game/%d/last_move", serverUrl, gameID)

	fmt.Println(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch last move, status code: %d", resp.StatusCode)
	}

	var move DbMove
	err = json.NewDecoder(resp.Body).Decode(&move)
	if err != nil {
		return nil, err
	}

	if move.GameID == 0 {
		return nil, nil
	}

	return &move, nil
}

func dbMoveToMove(dbmove *DbMove) chessboard.Move {
	from := &chessboard.Location{
		Rank: dbmove.MFrom % 8,
		File: dbmove.MFrom / 8,
	}
	to := &chessboard.Location{
		Rank: dbmove.MTo % 8,
		File: dbmove.MTo / 8,
	}
	return chessboard.Move{
		From: from,
		To:   to,
	}
}

func makeMove(gameID int, pieceID string, from *chessboard.Location, to *chessboard.Location, token string, serverUrl string) error {
	// Convert the locations to (x,y) tuples as expected by the server
	fromTuple := []int{from.File, from.Rank}
	toTuple := []int{to.File, to.Rank}
	fmt.Println("moving request from", fromTuple, "to", toTuple)

	// Create the request body
	moveData := map[string]interface{}{
		"piece_id": pieceID,
		"mfrom":    fromTuple,
		"mto":      toTuple,
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(moveData)
	if err != nil {
		return fmt.Errorf("error marshaling move data: %v", err)
	}

	// Create the request
	url := fmt.Sprintf("%s/_game/%d/move", serverUrl, gameID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

type DbUser struct {
	UID         int    `json:"uid"`
	GamesLost   int    `json:"games_lost"`
	GamesWon    int    `json:"games_won"`
	GamesDraw   int    `json:"games_draw"`
	WinStreak   int    `json:"win_streak"`
	LoseStreak  int    `json:"lose_streak"`
	CurrentElo  int    `json:"current_elo"`
	TotalGames  int    `json:"total_games"`
	Username    string `json:"username"`
	PeakElo     int    `json:"peak_elo"`
	PeakRank    int    `json:"peak_rank"`
	CurrentRank int    `json:"current_rank"`
}

func GetUser(token string, serverUrl string, uid int) (*DbUser, error) {
	url := fmt.Sprintf("%s/_user/%d", serverUrl, uid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication required")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error status %d", resp.StatusCode)
	}

	var user DbUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &user, nil
}

func GetAllUsers(token string, serverUrl string) ([]DbUser, error) {
	url := fmt.Sprintf("%s/_user/list", serverUrl)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication required")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error status %d", resp.StatusCode)
	}

	var users []DbUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return users, nil
}

func CreateUserSelector(token string, serverUrl string) (*widget.Select, []DbUser, error) {
	// Fetch all users first
	users, err := GetAllUsers(token, serverUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch users: %v", err)
	}

	// Create a map to store username to user mapping for easy lookup
	userMap := make(map[string]*DbUser)
	// Create slice of usernames for the dropdown
	usernames := make([]string, len(users))

	// Sort users by username for better UX
	sort.Slice(users, func(i, j int) bool {
		return users[i].Username < users[j].Username
	})

	// Populate the slices and map
	for i, user := range users {
		usernames[i] = user.Username
		userMap[user.Username] = &users[i]
	}

	// Create the dropdown
	selector := widget.NewSelect(usernames, nil)
	selector.PlaceHolder = "Select a user..."

	return selector, users, nil
}

func CreateNewGame(token string, serverUrl string, opponentUID int) (*DbGame, error) {
	url := fmt.Sprintf("%s/_game/new/%d", serverUrl, opponentUID)

	// Create the request
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
	var game DbGame
	if err := json.NewDecoder(resp.Body).Decode(&game); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &game, nil
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
	w := a.NewWindow("Choose Game")

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

	userSelector, userlist, err := CreateUserSelector(account.AuthToken, serverUrl)

	var selectedGame *DbGame

	// Add new game row
	gridELS = append(gridELS,
		widget.NewLabel("New:"),
		widget.NewLabel(account.Cred.Username), // Tournament placeholder
		widget.NewLabel(""),
		userSelector,
		widget.NewLabel(""),
		widget.NewButtonWithIcon("Select Opponent", theme.AccountIcon(), func() {
			selectedUsername := userSelector.Selected
			fmt.Println("Selected User:", selectedUsername)
			if selectedUsername == "" {
				dialog.ShowInformation("No user selected.", "You must select a user.", w)
				return
			}

			var selectedUserObj *DbUser
			for _, user := range userlist {
				if user.Username == selectedUsername {
					selectedUserObj = &user
					break
				}
			}

			if selectedUserObj == nil {
				dialog.ShowInformation("Error", "Selected user not found.", w)
				return
			}

			fmt.Printf("Selected user ID: %d\n", selectedUserObj.UID)

			selectedGame, err = CreateNewGame(account.AuthToken, serverUrl, selectedUserObj.UID)
			if err != nil {
				dialog.ShowError(err, w)
			}
			desiredGameIDToPlay = selectedGame.GameID

			w.Close()
		}),
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

	if selectedGame == nil {
		// Find the selected game
		for _, game := range games {
			if game.GameID == desiredGameIDToPlay {
				gameCopy := game
				selectedGame = &gameCopy
				break
			}
		}
	}

	if selectedGame == nil && desiredGameIDToPlay == 0 {
		return true
	} else if selectedGame == nil {
		panic("No game found with ID:" + strconv.Itoa(desiredGameIDToPlay))
	}

	fmt.Printf("White: %s, Black: %s\n", selectedGame.WhiteName, selectedGame.BlackName)

	dbmoves := selectedGame.Moves
	sort.Slice(dbmoves, func(i, j int) bool {
		return dbmoves[i].MIndex < dbmoves[j].MIndex
	})

	gameApp := app.New()
	gameWindow := gameApp.NewWindow("Online Game")

	viewingHistorical := atomic.Bool{}
	viewedMove := atomic.Int32{}

	moves := make([]chessboard.Move, 0)

	board := chessboard.NewChessBoard()

	playingText := widget.NewLabel("White's game...")
	updatePlayingText := func(isBlackGame bool) {
		if isBlackGame {
			playingText.SetText("Black's game...")
		} else {
			playingText.SetText("White's game...")
		}
	}
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

	for _, dbmove := range dbmoves {
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
	isBlackTurn := selectedGame.Turn == "B"
	ourTurn := isBlack == isBlackTurn

	updatePlayingText(isBlackTurn)

	fmt.Println("turn", selectedGame.Turn)

	gameloopRunning := atomic.Bool{}
	gameloopRunning.Store(true)

	go func() {
		movesWeMade := make([]chessboard.Move, 0)
		movesTheyMade := make([]chessboard.Move, 0)

		lastFromRank := atomic.Int32{}
		lastToRank := atomic.Int32{}
		lastFromFile := atomic.Int32{}
		lastToFile := atomic.Int32{}

		for {

			if !gameloopRunning.Load() {
				return
			}

			var move chessboard.Move
			if !ourTurn {

				time.Sleep(1 * time.Second)

				ldbm, err := fetchLastMove(selectedGame.GameID, account.AuthToken, serverUrl)
				if err != nil {
					fyne.Do(func() {
						dialog.ShowInformation("Error checking last move", err.Error(), gameWindow)
						gameloopRunning.Store(false)
						gameWindow.Close()
					})
					continue
				}
				fmt.Println("last move", ldbm)

				if ldbm == nil {
					continue
				}

				//check if we already have last move
				old := false
				for _, dbmove := range dbmoves {
					fmt.Println(dbmove.MIndex, ldbm.MIndex)
					if dbmove.MIndex == ldbm.MIndex {
						old = true
					}
				}

				if old {
					continue
				}

				dbmoves = append(dbmoves, *ldbm)

				move = dbMoveToMove(ldbm)

				movesTheyMade = append(movesTheyMade, move)

				fmt.Println(*ldbm)
				ourTurn = true

			} else {
				if len(movesWeMade) > 0 &&
					movesWeMade[len(movesWeMade)-1].From.File == moves[len(moves)-1].From.File &&
					movesWeMade[len(movesWeMade)-1].From.Rank == moves[len(moves)-1].From.Rank &&
					movesWeMade[len(movesWeMade)-1].From.Rank == moves[len(moves)-1].From.Rank &&
					movesWeMade[len(movesWeMade)-1].From.File == moves[len(moves)-1].From.File {
					ourTurn = false
					continue

				}

				var startPosChan chan *chessboard.Location
				var endPosChan chan *chessboard.Location
				var startPos *chessboard.Location
				var endPos *chessboard.Location
				//cancel op is selecting the origina tile
				for startPos == nil ||
					endPos == nil ||
					(startPos.Rank == endPos.Rank &&
						startPos.File == endPos.File) {

					//fall back for when no options are there, nil chanel will be returned
					for ok := true; ok; ok = endPosChan == nil {

						fyne.DoAndWait(func() { startPosChan = board.PrepareForMove(isBlack, false) })

						startPos = <-startPosChan
						fmt.Println(startPos, "startPos")
						fyne.DoAndWait(func() { endPosChan = board.MoveChooser(startPos.Rank, startPos.File) })
						fmt.Println(endPosChan)
					}
					endPos = <-endPosChan

					if startPos.Rank == endPos.Rank &&
						startPos.File == endPos.File {
						fmt.Println("Move is no-op, allowing user to chose piece to move again")
					}
				}

				move = chessboard.Move{From: startPos, To: endPos}
				movesWeMade = append(movesWeMade, move)
				ourTurn = false

				fmt.Println(lastFromFile.Load() == int32(move.From.File))
				fmt.Println(lastToFile.Load() == int32(move.To.File))
				fmt.Println(lastFromRank.Load() == int32(move.From.Rank))
				fmt.Println(lastToRank.Load() == int32(move.To.Rank))

				if lastFromFile.Load() == int32(move.From.File) &&
					lastToFile.Load() == int32(move.To.File) &&
					lastFromRank.Load() == int32(move.From.Rank) &&
					lastToRank.Load() == int32(move.To.Rank) {
					continue
				}
				lastFromFile.Store(int32(move.From.File))
				lastFromRank.Store(int32(move.From.Rank))
				lastToFile.Store(int32(move.To.File))
				lastToRank.Store(int32(move.To.Rank))

				go func() {

					fmt.Println("in goroutine")
					err = makeMove(
						selectedGame.GameID,
						board.Tiles[move.To.Rank][move.To.File].Piece.PieceType,
						move.From,
						move.To,
						account.AuthToken,
						serverUrl,
					)
					if err != nil {
						fyne.Do(func() {
							dialog.ShowInformation("Error making move", err.Error(), gameWindow)
						})
					}
				}()

			}

			isBlackTurn = isBlack == ourTurn
			fyne.Do(func() {
				updatePlayingText(isBlackTurn)
			})

			updateMoveStore := true
			if viewingHistorical.Load() {
				fmt.Println("Not updating grid as we are viewing historical move")
			} else {
				fyne.DoAndWait(func() { updateMoveStore = board.MovePiece(move.From, move.To, false) })
			}

			if updateMoveStore {
				moves = append(moves, move)
				fmt.Println(len(moves), "moves", moves)
				viewedMove.Store(int32(len(moves)))
			}
			updateViewingText()

			fmt.Println("our turn", ourTurn)

		}
	}()

	//board.PrepareForMove()

	gameWindow.ShowAndRun()

	gameloopRunning.Store(false)

	return true
}

func Profile(account AccountData, serverUrl string) {
	profileApp := app.New()
	profileWindow := profileApp.NewWindow("User Profile")

	userSelector, users, err := CreateUserSelector(account.AuthToken, serverUrl)
	if err != nil {
		dialog.ShowError(err, profileWindow)
	}

	var profileForm *fyne.Container
	var selectedUserObj *DbUser

	userSelector.OnChanged = func(s string) {
		for _, user := range users {
			if user.Username == s {
				selectedUserObj = &user
				break
			}
		}

		profileForm.RemoveAll()
		profileForm.Add(widget.NewLabel("Profile"))
		profileForm.Add(userSelector)
		profileForm.Add(widget.NewLabel("Username:"))
		profileForm.Add(widget.NewLabel(selectedUserObj.Username))
		profileForm.Add(widget.NewLabel("User ID:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.UID)))
		profileForm.Add(widget.NewLabel("Current Elo:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.CurrentElo)))
		profileForm.Add(widget.NewLabel("Peak Elo:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.PeakElo)))
		profileForm.Add(widget.NewLabel("Current Rank:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.CurrentRank)))
		profileForm.Add(widget.NewLabel("Peak Rank:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.PeakRank)))
		profileForm.Add(widget.NewLabel("Games Won:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.GamesWon)))
		profileForm.Add(widget.NewLabel("Games Lost:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.GamesLost)))
		profileForm.Add(widget.NewLabel("Games Draw:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.GamesDraw)))
		profileForm.Add(widget.NewLabel("Total Games:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.TotalGames)))
		profileForm.Add(widget.NewLabel("Win Streak:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.WinStreak)))
		profileForm.Add(widget.NewLabel("Lose Streak:"))
		profileForm.Add(widget.NewLabel(fmt.Sprintf("%d", selectedUserObj.LoseStreak)))
	}

	profileForm = container.NewVBox(
		widget.NewLabel("Profile"),
		userSelector,
		widget.NewLabel("Username:"),
		widget.NewLabel(""),
		widget.NewLabel("User ID:"),
		widget.NewLabel(""),
		widget.NewLabel("Current Elo:"),
		widget.NewLabel(""),
		widget.NewLabel("Peak Elo:"),
		widget.NewLabel(""),
		widget.NewLabel("Current Rank:"),
		widget.NewLabel(""),
		widget.NewLabel("Peak Rank:"),
		widget.NewLabel(""),
		widget.NewLabel("Games Won:"),
		widget.NewLabel(""),
		widget.NewLabel("Games Lost:"),
		widget.NewLabel(""),
		widget.NewLabel("Games Draw:"),
		widget.NewLabel(""),
		widget.NewLabel("Total Games:"),
		widget.NewLabel(""),
		widget.NewLabel("Win Streak:"),
		widget.NewLabel(""),
		widget.NewLabel("Lose Streak:"),
		widget.NewLabel(""),
	)

	content := container.NewVScroll(profileForm)
	profileWindow.SetContent(content)
	profileWindow.Resize(fyne.NewSize(300, 400))
	profileWindow.ShowAndRun()
}
