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

func NewDefaultPieceForPosition(x int, y int) *ChessPiece {
	newPiece := ChessPiece{}

	if y > 1 && y < 6 {
		return &newPiece
	}

	//top side is black
	newPiece.black = y > 3

	//outer pieces are always pawns
	if y == 1 || y == 6 {
		newPiece.piece = "pawn"
	} else {
		// https://www.regencychess.co.uk/images/how-to-set-up-a-chessboard/how-to-set-up-a-chessboard-7.jpg
		switch x {
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

	//get the image asset
	var colorName string
	if newPiece.black {
		colorName = "black"
	} else {
		colorName = "white"
	}
	newPiece.imageEL = canvas.NewImageFromFile("./assets/pieces/" + colorName + "/" + newPiece.piece + ".png")

	//new instance of newPiece
	return &newPiece
}

type ChessTile struct {
	piece  *ChessPiece
	button *widget.Button
	uiEL   *fyne.Container
}

func initChessTileAtPos(x int, y int) *ChessTile {

	//init things
	piece := NewDefaultPieceForPosition(x, y)
	button := widget.NewButton("", func() {
		fmt.Println("button clicked", x, y)
	})
	//uiEL := container.NewVBox(piece.imageEL, button)
	var uiEL *fyne.Container
	if piece.imageEL == nil {
		uiEL = container.NewVBox(button)
	} else {
		uiEL = container.NewHBox(piece.imageEL, button)
	}

	//create obj and return
	return &ChessTile{
		piece:  piece,
		button: button,
		uiEL:   uiEL,
	}
}

type ChessBoard struct {
	grid  *fyne.Container
	tiles [8][8]*ChessTile
}

func NewChessBoard() *ChessBoard {
	board := ChessBoard{}

	//1d array for grid layout
	tiles := make([]*fyne.CanvasObject, 64)

	iter := 0
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {

			//init tile
			tile := initChessTileAtPos(x, y)

			//2d array for managing easier
			board.tiles[x][y] = tile

			//1d array for grid layout
			tiles[iter] = tile.uiEL
			iter++
		}
	}

	grid := container.New(layout.NewGridLayout(8), tiles...)

}
