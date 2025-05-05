package gameModes

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/jjj333-p/chess-fe-go/chessboard"
)

func PracticeGame() bool {
	gameApp := app.New()
	gameWindow := gameApp.NewWindow("Practice Game")

	board := chessboard.NewChessBoard()

	gameWindow.SetContent(board.Grid)

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

			fyne.DoAndWait(func() { board.MovePiece(startPos, endPos) })

			fyne.DoAndWait(board.Grid.Refresh)
		}
	}()

	gameWindow.ShowAndRun()
	return true
}
