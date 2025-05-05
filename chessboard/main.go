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
	Black     bool
	PieceType string
	ImageEL   *canvas.Image
}

type Location struct {
	Rank int
	File int
}

type Move struct {
	From *Location
	To   *Location
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
	newPiece.Black = rank > 4

	// Pawns on ranks 2 and 7 (index 1 and 6)
	if rank == 1 || rank == 6 {
		newPiece.PieceType = "pawn"
	} else {
		// Back Rank pieces (ranks 1 and 8)
		switch file {
		case 0, 7:
			newPiece.PieceType = "rook"
		case 1, 6:
			newPiece.PieceType = "knight"
		case 2, 5:
			newPiece.PieceType = "bishop"
		case 3:
			newPiece.PieceType = "queen"
		case 4:
			newPiece.PieceType = "king"
		}
	}

	colorName := "white"
	if newPiece.Black {
		colorName = "black"
	}
	newPiece.ImageEL = canvas.NewImageFromFile("./assets/pieces/" + colorName + "/" + newPiece.PieceType + ".png")
	newPiece.ImageEL.FillMode = canvas.ImageFillContain
	newPiece.ImageEL.SetMinSize(fyne.NewSize(40, 40))

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
	fmt.Println("Assembling", self.Name, "with image el", self.Piece.ImageEL, " - refresh is ", refresh)

	//if ui doesnt exist create it, else remove all items and insert again
	if self.uiTop == nil {
		//show piece if it is there
		if self.Piece.ImageEL == nil {
			self.uiTop = container.NewVBox(layout.NewSpacer(), self.Button)
		} else {
			self.uiTop = container.NewVBox(layout.NewSpacer(), self.Piece.ImageEL, self.Button)
		}
	} else {
		self.uiTop.RemoveAll()
		if self.Piece.ImageEL == nil {
			self.uiTop.Add(layout.NewSpacer())
			self.uiTop.Add(self.Button)
		} else {
			self.uiTop.Add(layout.NewSpacer())
			self.uiTop.Add(self.Piece.ImageEL)
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
	Grid    *fyne.Container
	Tiles   [8][8]*ChessTile
	discard []*ChessPiece
}

// DisableAllBtn disables all buttons on the board. Simple as.
func (self *ChessBoard) DisableAllBtn() {
	for _, fileSlice := range self.Tiles {
		for _, tile := range fileSlice {
			fyne.Do(func() {
				if !tile.Button.Disabled() {
					tile.Button.SetText(tile.Name)
					tile.Button.Disable()
					tile.Button.Refresh()
				}
			})
			tile.moveChan = nil
		}
	}
}

/*
UiEls returns a flat slice of all ui elements.
Intended to be used for putting into the grid layout.
`reverse` indicates basically if the board should be flipped around
such as if the Black side is playing.
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
				//check that the PieceType is white (we can play it) and defined
				if tile.Piece.Black && tile.Piece.PieceType != "" {
					tile.moveChan = &innerMoveChan
					tile.Button.Enable()
					tile.Button.SetText("Move " + tile.Piece.PieceType)
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
				//check that the PieceType is white (we can play it) and defined
				if !tile.Piece.Black && tile.Piece.PieceType != "" {
					tile.moveChan = &innerMoveChan
					tile.Button.Enable()
					tile.Button.SetText("Move " + tile.Piece.PieceType)
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

	//cancel option
	tile.moveChan = &innerMoveChan
	tile.Button.SetText("Cancel")
	tile.Button.Enable()

	fmt.Println(tile.Piece.PieceType)
	moveableSpots := 0

	enableMoveButton := func(desiredTile *ChessTile) {
		moveableSpots += 1
		desiredTile.moveChan = &innerMoveChan
		desiredTile.Button.SetText("Move here")
		desiredTile.Button.Enable()
	}

	searchList := func(locationsToLook *[8]Location) {
		for _, l := range locationsToLook {
			if l.Rank < 0 || l.Rank > 7 {
				fmt.Println(l, "rank is out of range")
				continue
			}
			if l.File < 0 || l.File > 7 {
				fmt.Println(l, "file is out of range")
				continue
			}

			targetTile := self.Tiles[l.Rank][l.File]
			if targetTile.Piece.PieceType == "" ||
				targetTile.Piece.Black != self.Tiles[rank][file].Piece.Black {
				enableMoveButton(targetTile)
			} else {
				fmt.Println("Knight cannot take its own color at", l)
			}
		}
	}

	rookLogic := func() {
		//vertical up
		for targetRank := rank + 1; targetRank < 8; targetRank++ {
			targetTile := self.Tiles[targetRank][file]
			if targetTile.Piece.PieceType == "" {
				fmt.Println(targetTile.Name, "is empty, opening and continuing search.")
				enableMoveButton(targetTile)
			} else {
				if targetTile.Piece.Black == self.Tiles[rank][file].Piece.Black {
					fmt.Println(targetTile.Name, "Colors match, not opening and stopping search.")
				} else {
					fmt.Println(targetTile.Name, "Colors don't match, opening and stopping search.")
					enableMoveButton(targetTile)
				}
				break
			}
		}

		//vertical down
		for targetRank := rank - 1; targetRank > -1; targetRank-- {
			targetTile := self.Tiles[targetRank][file]
			if targetTile.Piece.PieceType == "" {
				fmt.Println(targetTile.Name, "is empty, opening and continuing search.")
				enableMoveButton(targetTile)
			} else {
				if targetTile.Piece.Black == self.Tiles[rank][file].Piece.Black {
					fmt.Println(targetTile.Name, "Colors match, not opening and stopping search.")
				} else {
					fmt.Println(targetTile.Name, "Colors don't match, opening and stopping search.")
					enableMoveButton(targetTile)
				}
				break
			}
		}

		//horizontal right
		for targetFile := file + 1; targetFile < 8; targetFile++ {
			targetTile := self.Tiles[rank][targetFile]
			if targetTile.Piece.PieceType == "" {
				fmt.Println(targetTile.Name, "is empty, opening and continuing search.")
				enableMoveButton(targetTile)
			} else {
				if targetTile.Piece.Black == self.Tiles[rank][file].Piece.Black {
					fmt.Println(targetTile.Name, "Colors match, not opening and stopping search.")
				} else {
					fmt.Println(targetTile.Name, "Colors don't match, opening and stopping search.")
					enableMoveButton(targetTile)
				}
				break
			}
		}

		//horizontal left
		for targetFile := file - 1; targetFile > -1; targetFile-- {
			targetTile := self.Tiles[rank][targetFile]
			if targetTile.Piece.PieceType == "" {
				fmt.Println(targetTile.Name, "is empty, opening and continuing search.")
				enableMoveButton(targetTile)
			} else {
				if targetTile.Piece.Black == self.Tiles[rank][file].Piece.Black {
					fmt.Println(targetTile.Name, "Colors match, not opening and stopping search.")
				} else {
					fmt.Println(targetTile.Name, "Colors don't match, opening and stopping search.")
					enableMoveButton(targetTile)
				}
				break
			}
		}
	}
	bishopLogic := func() {
		//up and to the right
		for increment := 1; increment < 7; increment++ {
			fmt.Println(increment)
			targetRank := rank + increment
			targetFile := file + increment
			if targetRank < 0 || targetRank > 7 {
				fmt.Println(targetRank, "rank is out of range, stopping search.")
				break
			}
			if targetFile < 0 || targetFile > 7 {
				fmt.Println(targetFile, "file is out of range, stopping search.")
				break
			}
			targetTile := self.Tiles[targetRank][targetFile]

			if targetTile.Piece.PieceType == "" {
				fmt.Println(targetTile.Name, "is empty, opening and continuing search.")
				enableMoveButton(targetTile)
			} else {
				if targetTile.Piece.Black == self.Tiles[rank][file].Piece.Black {
					fmt.Println(targetTile.Name, "Colors match, not opening and stopping search.")
				} else {
					fmt.Println(targetTile.Name, "Colors don't match, opening and stopping search.")
					enableMoveButton(targetTile)
				}
				break
			}
		}

		// up and to the left
		for increment := 1; increment < 7; increment++ {
			targetRank := rank + increment
			targetFile := file - increment
			if targetRank < 0 || targetRank > 7 {
				fmt.Println(targetRank, "rank is out of range, stopping search.")
				break
			}

			if targetFile < 0 || targetFile > 7 {
				fmt.Println(targetFile, "file is out of range, stopping search.")
				break
			}
			targetTile := self.Tiles[targetRank][targetFile]

			if targetTile.Piece.PieceType == "" {
				fmt.Println(targetTile.Name, "is empty, opening and continuing search.")
				enableMoveButton(targetTile)
			} else {
				if targetTile.Piece.Black == self.Tiles[rank][file].Piece.Black {
					fmt.Println(targetTile.Name, "Colors match, not opening and stopping search.")
				} else {
					fmt.Println(targetTile.Name, "Colors don't match, opening and stopping search.")
					enableMoveButton(targetTile)
				}
				break
			}
		}

		//down and to the right
		for increment := 1; increment < 7; increment++ {
			targetRank := rank - increment
			targetFile := file + increment
			if targetRank < 0 || targetRank > 7 {
				fmt.Println(targetRank, "rank is out of range, stopping search.")
				break
			}
			if targetFile < 0 || targetFile > 7 {
				fmt.Println(targetFile, "file is out of range, stopping search.")
				break
			}

			targetTile := self.Tiles[targetRank][targetFile]

			if targetTile.Piece.PieceType == "" {
				fmt.Println(targetTile.Name, "is empty, opening and continuing search.")
				enableMoveButton(targetTile)
			} else {
				if targetTile.Piece.Black == self.Tiles[rank][file].Piece.Black {
					fmt.Println(targetTile.Name, "Colors match, not opening and stopping search.")
				} else {
					fmt.Println(targetTile.Name, "Colors don't match, opening and stopping search.")
					enableMoveButton(targetTile)
				}
				break
			}
		}

		//down and to the left
		for increment := 1; increment < 7; increment++ {
			targetRank := rank - increment
			targetFile := file - increment
			if targetRank < 0 || targetRank > 7 {
				fmt.Println(targetRank, "rank is out of range, stopping search.")
				break
			}
			if targetFile < 0 || targetFile > 7 {
				fmt.Println(targetFile, "file is out of range, stopping search.")
				break
			}
			targetTile := self.Tiles[targetRank][targetFile]

			if targetTile.Piece.PieceType == "" {
				fmt.Println(targetTile.Name, "is empty, opening and continuing search.")
				enableMoveButton(targetTile)
			} else {
				if targetTile.Piece.Black == self.Tiles[rank][file].Piece.Black {
					fmt.Println(targetTile.Name, "Colors match, not opening and stopping search.")
				} else {
					fmt.Println(targetTile.Name, "Colors don't match, opening and stopping search.")
					enableMoveButton(targetTile)
				}
				break
			}
		}
	}

	switch tile.Piece.PieceType {
	case "pawn":
		var targetRank int
		if tile.Piece.Black {
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

		if fowardMoveTile.Piece.PieceType == "" {
			fmt.Println(fowardMoveTile.Name)
			enableMoveButton(fowardMoveTile)
		} else {
			fmt.Println("Cannot move pawn foward as a", fowardMoveTile.Piece.PieceType, "is in the way.")
		}

		if file < 7 {
			rightTakeTile := self.Tiles[targetRank][file+1]
			if rightTakeTile.Piece.PieceType == "" {
				fmt.Println("Cannot take right with pawn as there is no piece to take.")
			} else if rightTakeTile.Piece.Black == tile.Piece.Black {
				fmt.Println("Cannot take right with pawn as we cannot take our own piece.")
			} else {
				fmt.Println(rightTakeTile.Name)
				enableMoveButton(rightTakeTile)
			}
		} else {
			fmt.Println("Cannot take right as that is off the board")
		}

		if file > 0 {
			leftTakeTile := self.Tiles[targetRank][file-1]
			if leftTakeTile.Piece.PieceType == "" {
				fmt.Println("Cannot take left with pawn as there is no piece to take.")
			} else if leftTakeTile.Piece.Black == tile.Piece.Black {
				fmt.Println("Cannot take right with pawn as we cannot take our own piece.")
			} else {
				fmt.Println(leftTakeTile.Name)
				enableMoveButton(leftTakeTile)
			}
		}

		fmt.Println("The pawn can move", moveableSpots, "spaces.")
	case "rook":
		rookLogic()
	case "knight":
		// easier to statically define these and iterate than figure out a working algorithm
		locationsToLook := &[8]Location{
			{rank + 2, file + 1},
			{rank + 1, file + 2},
			{rank - 2, file + 1},
			{rank - 1, file + 2},
			{rank - 2, file - 1},
			{rank - 1, file - 2},
			{rank + 2, file - 1},
			{rank + 1, file - 2},
		}

		searchList(locationsToLook)

	case "bishop":
		bishopLogic()
	case "queen":
		rookLogic()
		bishopLogic()
	case "king":
		// easier to statically define these and iterate than figure out a working algorithm
		locationsToLook := &[8]Location{
			{rank + 1, file + 1},
			{rank + 1, file},
			{rank + 1, file - 1},
			{rank, file + 1},
			{rank, file - 1},
			{rank - 1, file - 1},
			{rank - 1, file},
			{rank - 1, file + 1},
		}

		searchList(locationsToLook)
	}

	if moveableSpots > 0 {
		return moveChan
	} else {
		return nil
	}
}

func (self *ChessBoard) MovePiece(from *Location, to *Location, reverse bool) {

	fmt.Println("moving from", from, "to", to)

	var last *ChessPiece
	if len(self.discard) > 0 {
		last = self.discard[len(self.discard)-1]
	} else {
		last = &ChessPiece{}
	}

	if self.Tiles[from.Rank][from.File].Piece.PieceType == "" &&
		(last.PieceType == "" || !reverse) {
		fmt.Println("not moving empty element (this should not happen)")
		return
	}

	hold := self.Tiles[to.Rank][to.File].Piece

	// move to new position
	self.Tiles[to.Rank][to.File].Piece = self.Tiles[from.Rank][from.File].Piece
	self.Tiles[from.Rank][from.File].Piece = &ChessPiece{}

	//pawn becomes queen at back
	if self.Tiles[to.Rank][to.File].Piece.PieceType == "pawn" && to.Rank == 7 {
		var colorName string
		if self.Tiles[to.Rank][to.File].Piece.Black {
			colorName = "black"
		} else {
			colorName = "white"
		}
		imageEL := canvas.NewImageFromFile("./assets/pieces/" + colorName + "/queen.png")
		imageEL.FillMode = canvas.ImageFillContain
		imageEL.SetMinSize(fyne.NewSize(40, 40))
		self.Tiles[to.Rank][to.File].Piece = &ChessPiece{
			Black:     self.Tiles[to.Rank][to.File].Piece.Black,
			PieceType: "queen",
			ImageEL:   imageEL,
		}
	}

	//update ui
	self.Tiles[from.Rank][from.File].AssembleUI(true)
	self.Tiles[to.Rank][to.File].AssembleUI(true)

	if reverse {
		self.Tiles[from.Rank][from.File].Piece = last
		self.Tiles[from.Rank][from.File].AssembleUI(true)
		self.discard = self.discard[:len(self.discard)-1]
	} else {
		self.discard = append(self.discard, hold)
	}

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
