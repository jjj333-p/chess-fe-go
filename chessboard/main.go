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

// NewDefaultPieceForPosition creates a new Piece based on position
// Rank is 0-7 (representing ranks 1-8)
// File is 0-7 (representing files a-h)
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
		// Back Rank pieces (ranks 1 and 8)
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
	Piece    *ChessPiece
	Name     string
	moveChan *chan *Location
	Button   *widget.Button
	UiEL     *fyne.Container
	BgColor  *canvas.Image
}

/*
initChessTileAtPos initializes a tile of the board at a position. This
implicitly creates a chess Piece to place based on starting positions.
This is not the code that creates the chess tiles though.
*/
func initChessTileAtPos(rank, file int) *ChessTile {
	newTile := &ChessTile{}
	newTile.Piece = NewDefaultPieceForPosition(rank, file)

	// Convert to chess notation for better debugging
	fileChar := rune('a' + file)
	rankNum := rank + 1

	newTile.Name = string(fileChar) + strconv.Itoa(rankNum)

	//move Button
	newTile.Button = widget.NewButton("", func() {
		fmt.Printf("Clicked square %c%d\n", fileChar, rankNum)
		if newTile.moveChan == nil {
			fmt.Println("Nothing to click")
		} else {
			*newTile.moveChan <- &Location{
				rank,
				file,
			}
		}
	})

	//start disabled
	newTile.Button.Disable()

	newTile.Button.SetText(newTile.Name)

	var top *fyne.Container
	if newTile.Piece.imageEL == nil {
		top = container.NewVBox(layout.NewSpacer(), newTile.Button)
	} else {
		top = container.NewVBox(layout.NewSpacer(), newTile.Piece.imageEL, newTile.Button)
	}

	// Determine square color: if Rank + File is even, it's a light square
	isLightSquare := (rank+file)%2 == 0
	if isLightSquare {
		newTile.BgColor = canvas.NewImageFromFile("./assets/bg/light.png")
	} else {
		newTile.BgColor = canvas.NewImageFromFile("./assets/bg/dark.png")
	}
	newTile.BgColor.SetMinSize(fyne.NewSize(70, 70))

	newTile.UiEL = container.NewStack(newTile.BgColor, top)

	return newTile
}

type ChessBoard struct {
	Grid  *fyne.Container
	Tiles [8][8]*ChessTile
}

func (self *ChessBoard) DisableAllBtn() {
	for _, fileSlice := range self.Tiles {
		for _, tile := range fileSlice {
			fyne.Do(func() {
				tile.Button.SetText(tile.Name)
				tile.Button.Disable()
				tile.Button.Refresh()
			})
		}
	}
}

func (self *ChessBoard) uiEls(reverse bool) []fyne.CanvasObject {
	uiTiles := make([]fyne.CanvasObject, 64)
	iter := 0

	if reverse {
		for rank := 0; rank < 8; rank++ {
			for file := 7; file >= 0; file-- {
				tile := self.Tiles[rank][file]
				uiTiles[iter] = tile.UiEL
				iter++
			}
		}
	} else {
		for rank := 7; rank >= 0; rank-- {
			for file := 0; file < 8; file++ {
				tile := self.Tiles[rank][file]
				uiTiles[iter] = tile.UiEL
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
func (self *ChessBoard) PrepareForMove(colorIsBlack bool, lastColorIsBlack bool) chan *Location {
	moveChan := make(chan *Location)
	innerMoveChan := make(chan *Location)
	go func(moveChan chan *Location, innerMoveChan chan *Location) {
		l := <-innerMoveChan
		self.DisableAllBtn()
		moveChan <- l
	}(moveChan, innerMoveChan)
	if colorIsBlack {
		//enable Button for pieces we can move
		for _, rankSlice := range self.Tiles {
			for _, tile := range rankSlice {
				//check that the pieceType is white (we can play it) and defined
				if tile.Piece.black && tile.Piece.pieceType != "" {
					tile.moveChan = &innerMoveChan
					tile.Button.Enable()
					tile.Button.SetText("Move " + tile.Piece.pieceType)
					tile.Button.Refresh()
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
		//enable Button for pieces we can move
		for _, rankSlice := range self.Tiles {
			for _, tile := range rankSlice {
				//check that the pieceType is white (we can play it) and defined
				if !tile.Piece.black && tile.Piece.pieceType != "" {
					tile.moveChan = &innerMoveChan
					tile.Button.Enable()
					tile.Button.SetText("Move " + tile.Piece.pieceType)
					tile.Button.Refresh()
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

	return moveChan

}

type Location struct {
	Rank int
	File int
}

func (self *ChessBoard) MoveChooser(rank int, file int) chan *Location {
	//loop until valid move
	self.DisableAllBtn()
	tile := self.Tiles[rank][file]

	moveChan := make(chan *Location)
	innerMoveChan := make(chan *Location)
	go func(moveChan chan *Location, innerMoveChan chan *Location) {
		l := <-innerMoveChan
		self.DisableAllBtn()
		moveChan <- l
	}(moveChan, innerMoveChan)

	switch tile.Piece.pieceType {
	case "pawn":
		var targetFile int
		if tile.Piece.black {
			targetFile = file - 1
		} else {
			targetFile = file + 1
		}

		//out of bounds
		if targetFile < 0 || targetFile > 7 {
			panic("Pawn should never reach the end without becoming a queen")
		}

		fowardMoveTile := self.Tiles[rank][targetFile]

		if fowardMoveTile.Piece.pieceType == "" {
			fowardMoveTile.Button.SetText("Move here")
			fowardMoveTile.Button.Enable()
			go func() {
				l := <-*fowardMoveTile.moveChan
				self.DisableAllBtn()
				moveChan <- l
			}()
		} else {
			return nil
		}

	}

	return moveChan
}

func NewChessBoard() *ChessBoard {
	board := ChessBoard{}
	uiTiles := make([]fyne.CanvasObject, 64)

	iter := 0
	// Start from Rank 7 (index) downward to match chess convention
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			tile := initChessTileAtPos(rank, file)
			board.Tiles[rank][file] = tile
			uiTiles[iter] = tile.UiEL
			iter++
		}
	}

	board.Grid = container.New(layout.NewGridLayout(8), uiTiles...)
	return &board
}
