package chessboard

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"strconv"
)

type ChessPiece struct {
	black   bool
	piece   string
	imageEL *canvas.Image
}

// NewDefaultPieceForPosition creates a new piece based on position
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
	piece   *ChessPiece
	button  *widget.Button
	uiEL    *fyne.Container
	bgColor *canvas.Image
}

func initChessTileAtPos(rank, file int) *ChessTile {
	piece := NewDefaultPieceForPosition(rank, file)

	// Convert to chess notation for better debugging
	fileChar := rune('a' + file)
	rankNum := rank + 1

	//move button
	button := widget.NewButton("", func() {
		fmt.Printf("Clicked square %c%d\n", fileChar, rankNum)
	})

	button.SetText(string(fileChar) + strconv.Itoa(rankNum))

	var top *fyne.Container
	if piece.imageEL == nil {
		top = container.NewVBox(layout.NewSpacer(), button)
	} else {
		top = container.NewVBox(layout.NewSpacer(), piece.imageEL, button)
	}

	// Determine square color: if rank + file is even, it's a light square
	isLightSquare := (rank+file)%2 == 0
	var bgColor *canvas.Image
	if isLightSquare {
		bgColor = canvas.NewImageFromFile("./assets/bg/light.png")
	} else {
		bgColor = canvas.NewImageFromFile("./assets/bg/dark.png")
	}
	bgColor.SetMinSize(fyne.NewSize(70, 70))

	return &ChessTile{
		piece:   piece,
		button:  button,
		uiEL:    container.NewStack(bgColor, top),
		bgColor: bgColor,
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
