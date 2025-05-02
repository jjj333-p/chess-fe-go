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
	black     bool
	pieceType string
	imageEL   *canvas.Image
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
		newPiece.pieceType = "pawn"
	} else {
		// Back rank pieces (ranks 1 and 8)
		switch file {
		case 0, 7:
			newPiece.pieceType = "rook"
		case 1, 6:
			newPiece.pieceType = "knight"
		case 2, 5:
			newPiece.pieceType = "bishop"
		case 3:
			newPiece.pieceType = "queen"
		case 4:
			newPiece.pieceType = "king"
		}
	}

	colorName := "white"
	if newPiece.black {
		colorName = "black"
	}
	newPiece.imageEL = canvas.NewImageFromFile("./assets/pieces/" + colorName + "/" + newPiece.pieceType + ".png")
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

/*
initChessTileAtPos initializes a tile of the board at a position. This
implicitly creates a chess piece to place based on starting positions.
This is not the code that creates the chess tiles though.
*/
func initChessTileAtPos(rank, file int) *ChessTile {
	piece := NewDefaultPieceForPosition(rank, file)

	// Convert to chess notation for better debugging
	fileChar := rune('a' + file)
	rankNum := rank + 1

	//move button
	button := widget.NewButton("", func() {
		fmt.Printf("Clicked square %c%d\n", fileChar, rankNum)
	})

	//start disabled
	button.Disable()

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

func (self *ChessBoard) uiEls(reverse bool) []fyne.CanvasObject {
	uiTiles := make([]fyne.CanvasObject, 64)
	iter := 0

	if reverse {
		for rank := 0; rank < 8; rank++ {
			for file := 7; file >= 0; file-- {
				tile := self.Tiles[rank][file]
				uiTiles[iter] = tile.uiEL
				iter++
			}
		}
	} else {
		for rank := 7; rank >= 0; rank-- {
			for file := 0; file < 8; file++ {
				tile := self.Tiles[rank][file]
				uiTiles[iter] = tile.uiEL
				iter++
			}
		}
	}

	return uiTiles
}

/*
PrepareForMove lays out the ui for the user of this client to make a move.
Takes `color` (string) which is the color of the user making the move on this client,
and `assumeFirst` (bool) which allows us to optimize
*/
func (self *ChessBoard) PrepareForMove(colorIsBlack bool, lastColorIsBlack bool) {
	if colorIsBlack {
		//enable button for pieces we can move
		for _, rankSlice := range self.Tiles {
			for _, tile := range rankSlice {
				//check that the pieceType is white (we can play it) and defined
				if tile.piece.black && tile.piece.pieceType != "" {
					tile.button.Enable()
					tile.button.SetText("Move " + tile.piece.pieceType)
					tile.button.Refresh()
				}
			}
		}

		//do we need to rotate it around for white to play
		if !lastColorIsBlack {
			self.Grid.RemoveAll()
			for _, e := range self.uiEls(true) {
				self.Grid.Add(e)
			}
			self.Grid.Refresh()
		}
	} else {
		//enable button for pieces we can move
		for _, rankSlice := range self.Tiles {
			for _, tile := range rankSlice {
				//check that the pieceType is white (we can play it) and defined
				if !tile.piece.black && tile.piece.pieceType != "" {
					tile.button.Enable()
					tile.button.SetText("Move " + tile.piece.pieceType)
					tile.button.Refresh()
				}
			}
		}

		//do we need to rotate it around for white to play
		if lastColorIsBlack {
			self.Grid.RemoveAll()
			for _, e := range self.uiEls(false) {
				self.Grid.Add(e)
			}
			self.Grid.Refresh()
		}
	}

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
