package chessboard

import "fyne.io/fyne/v2/canvas"

type ChessPiece struct {
	black   bool
	piece   string
	imageEL *canvas.Image
}

func newDefaultPieceForPosition(x int, y int) *ChessPiece {
	newPiece := ChessPiece{}
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
