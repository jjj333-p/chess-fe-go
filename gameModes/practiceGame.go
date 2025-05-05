package gameModes

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/jjj333-p/chess-fe-go/chessboard"
	"strconv"
	"sync/atomic"
)

func PracticeGame() bool {
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

	go func() {
		for blackPlayer := false; true; blackPlayer = !blackPlayer {
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

					fyne.DoAndWait(func() { startPosChan = board.PrepareForMove(blackPlayer, !blackPlayer) })

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

			fmt.Println(endPos.Rank, endPos.File)

			moves = append(moves, chessboard.Move{From: startPos, To: endPos})
			fmt.Println(len(moves), "moves", moves)

			if viewingHistorical.Load() {
				fmt.Println("Not updating grid as we are viewing historical move")
			} else {
				fyne.Do(func() { board.MovePiece(startPos, endPos) })
				viewedMove.Store(int32(len(moves)))
			}
			updateViewingText()
		}
	}()

	gameWindow.ShowAndRun()
	return true
}
