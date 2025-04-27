package chessboard

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ChessPiece struct {
	black   bool
	piece   string
	imageEL *canvas.Image
}

// rank is 0-7 (representing ranks 1-8)
// file is 0-7 (representing files a-h)
func NewDefaultPieceForPosition(rank, file int) *ChessPiece {
	newPiece := ChessPiece{}

	// Empty squares (ranks 3-6)
	if rank > 1 && rank < 6 {
		return &newPiece
	}

	// Black pieces on ranks 7-8 (index 6-7), White pieces on ranks 1-2 (index 0-1)
	newPiece.black = rank > 4

	// Pawns on ranks 2 and 7 (index 1 and 6)
	if rank == 1 || rank == 6 {
		newPiece.piece = "pawn"
	} else {
		// Back rank pieces (ranks 1 and 8)
		switch file {
		case 0, 7:
			newPiece.piece = "rook"
		case 1, 6:
			newPiece.piece = "knight"
		case 2, 5:
			newPiece.piece = "bishop"
		case 3:
			newPiece.piece = "queen"
		case 4:
			newPiece.piece = "king"
		}
	}

	colorName := "white"
	if newPiece.black {
		colorName = "black"
	}
	newPiece.imageEL = canvas.NewImageFromFile("./assets/pieces/" + colorName + "/" + newPiece.piece + ".png")
	newPiece.imageEL.FillMode = canvas.ImageFillContain
	newPiece.imageEL.SetMinSize(fyne.NewSize(40, 40))

	return &newPiece
}

type ChessTile struct {
	piece  *ChessPiece
	button *widget.Button
	uiEL   *fyne.Container
}

func initChessTileAtPos(rank, file int) *ChessTile {
	piece := NewDefaultPieceForPosition(rank, file)
	button := widget.NewButton("", func() {
		// Convert to chess notation for better debugging
		fileChar := rune('a' + file)
		rankNum := rank + 1
		fmt.Printf("Clicked square %c%d\n", fileChar, rankNum)
	})

	var uiEL *fyne.Container
	if piece.imageEL == nil {
		uiEL = container.NewVBox(button)
	} else {
		uiEL = container.NewVBox(piece.imageEL, button)
	}

	return &ChessTile{
		piece:  piece,
		button: button,
		uiEL:   uiEL,
	}
}

type ChessBoard struct {
	Grid  *fyne.Container
	Tiles [8][8]*ChessTile
}

func NewChessBoard() *ChessBoard {
	board := ChessBoard{}
	uiTiles := make([]fyne.CanvasObject, 64)

	iter := 0
	// Start from rank 7 (index) downward to match chess convention
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			tile := initChessTileAtPos(rank, file)
			board.Tiles[rank][file] = tile
			uiTiles[iter] = tile.uiEL
			iter++
		}
	}

	board.Grid = container.New(layout.NewGridLayout(8), uiTiles...)
	return &board
}
