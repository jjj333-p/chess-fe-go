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

type Location struct {
	Rank int
	File int
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
	uiTop    *fyne.Container
	UiEL     *fyne.Container
	BgColor  *canvas.Image
}

/*
AssembleUI tells the chess tile to (re)form the ui.
This can be used to create the initial tile UI
or if one already exists, it will update the piece element
*/
func (self *ChessTile) AssembleUI(refresh bool) {
	fmt.Println("Assembling", self.Name, "with image el", self.Piece.imageEL, " - refresh is ", refresh)

	//if ui doesnt exist create it, else remove all items and insert again
	if self.uiTop == nil {
		//show piece if it is there
		if self.Piece.imageEL == nil {
			self.uiTop = container.NewVBox(layout.NewSpacer(), self.Button)
		} else {
			self.uiTop = container.NewVBox(layout.NewSpacer(), self.Piece.imageEL, self.Button)
		}
	} else {
		self.uiTop.RemoveAll()
		if self.Piece.imageEL == nil {
			self.uiTop.Add(layout.NewSpacer())
			self.uiTop.Add(self.Button)
		} else {
			self.uiTop.Add(layout.NewSpacer())
			self.uiTop.Add(self.Piece.imageEL)
			self.uiTop.Add(self.Button)
		}
	}

	/*
		if theres no ui element, it needs to be created and used, and that will refresh it
		else we should refresh it.
	*/
	if refresh {
		self.UiEL.Refresh()
	} else {
		self.UiEL = container.NewStack(self.BgColor, self.uiTop)
	}
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

	//Name for easy setting later
	newTile.Name = string(fileChar) + strconv.Itoa(rankNum)

	//move Button
	newTile.Button = widget.NewButton("", func() {
		fmt.Printf("Clicked square %c%d\n", fileChar, rankNum)

		//moveChan of the tile object is basically used as a radio to broadcast a button press over
		if newTile.moveChan == nil {
			//fmt.Println("Nothing to click")
			panic("There is no action behind this button. Why????")
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

	// Determine square color: if Rank + File is even, it's a light square
	isLightSquare := (rank+file)%2 == 0
	if isLightSquare {
		newTile.BgColor = canvas.NewImageFromFile("./assets/bg/light.png")
	} else {
		newTile.BgColor = canvas.NewImageFromFile("./assets/bg/dark.png")
	}
	newTile.BgColor.SetMinSize(fyne.NewSize(70, 70))

	newTile.AssembleUI(false)

	return newTile
}

type ChessBoard struct {
	Grid  *fyne.Container
	Tiles [8][8]*ChessTile
}

// DisableAllBtn disables all buttons on the board. Simple as.
func (self *ChessBoard) DisableAllBtn() {
	for _, fileSlice := range self.Tiles {
		for _, tile := range fileSlice {
			fyne.Do(func() {
				fmt.Println("Disabling tile", tile.Name)
				tile.Button.SetText(tile.Name)
				tile.Button.Disable()
				tile.Button.Refresh()
				fmt.Println("unlocked tile", tile.Name)
			})
			tile.moveChan = nil
		}
	}
}

/*
UiEls returns a flat slice of all ui elements.
Intended to be used for putting into the grid layout.
`reverse` indicates basically if the board should be flipped around
such as if the black side is playing.
*/
func (self *ChessBoard) UiEls(reverse bool) []fyne.CanvasObject {
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

	//concurrent thread to await a selection being returned, disable buttons, and pass it on
	go func(moveChan chan *Location, innerMoveChan chan *Location) {
		l := <-innerMoveChan
		self.DisableAllBtn()
		moveChan <- l
	}(moveChan, innerMoveChan)

	//logic to enable all pieces that are of the playing color
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
			for _, e := range self.UiEls(true) {
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
			for _, e := range self.UiEls(false) {
				self.Grid.Add(e)
			}
			self.Grid.Refresh()
		}
	}

	return moveChan

}

/*
MoveChooser is the successor to PrepareForMove() which lays out a selection of possible moves.
A nil chan indicates that no adequate move was available and the selection process should
be retried.
*/
func (self *ChessBoard) MoveChooser(rank int, file int) chan *Location {

	tile := self.Tiles[rank][file]

	moveChan := make(chan *Location)
	innerMoveChan := make(chan *Location)

	//thread to wait a selection to be made, then disable all buttons and return it
	go func(moveChan chan *Location, innerMoveChan chan *Location) {
		l := <-innerMoveChan
		fmt.Println(l, "move location from within move chooser")
		self.DisableAllBtn()
		moveChan <- l
		fmt.Println("done moving chooser")
	}(moveChan, innerMoveChan)

	fmt.Println(tile.Piece.pieceType)

	switch tile.Piece.pieceType {
	case "pawn":
		var targetRank int
		if tile.Piece.black {
			targetRank = rank - 1
		} else {
			targetRank = rank + 1
		}

		fmt.Println(targetRank)

		//out of bounds
		if targetRank < 0 || targetRank > 7 {
			panic("Pawn should never reach the end without becoming a queen")
		}

		fowardMoveTile := self.Tiles[targetRank][file]

		if fowardMoveTile.Piece.pieceType == "" {
			fmt.Println(fowardMoveTile.Name)
			fowardMoveTile.moveChan = &innerMoveChan
			fowardMoveTile.Button.SetText("Move here")
			fowardMoveTile.Button.Enable()
		} else {
			panic(fowardMoveTile.Piece.pieceType)
		}

	}

	return moveChan
}

func (self *ChessBoard) MovePiece(from *Location, to *Location) {
	fmt.Println("moving from", from, "to", to)

	// move to new position
	self.Tiles[to.Rank][to.File].Piece = self.Tiles[from.Rank][from.File].Piece
	self.Tiles[from.Rank][from.File].Piece = &ChessPiece{}

	//update ui
	self.Tiles[from.Rank][from.File].AssembleUI(true)
	self.Tiles[to.Rank][to.File].AssembleUI(true)

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
