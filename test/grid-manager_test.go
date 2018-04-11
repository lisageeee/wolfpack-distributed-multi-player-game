package test

import "testing"
import "../geometry"
import (
	"../shared"
	"fmt"
)


func setup() (geometry.GridManager) {
	gs := shared.InitialGameSettings{3000, 3000,
	[]shared.Coord{{1,1}, {10, 90}, {23, 99} }, 200}
	gm := geometry.CreateNewGridManager(gs)
	return gm
}

func TestInBoundsMiddle(t *testing.T) {
	gm := setup()
	inBoundsCoordMiddle := shared.Coord{50,50}
	response := gm.IsInBounds(inBoundsCoordMiddle)
	if !response {
		t.Fail()
	}
}

func TestInBoundsEdgeX(t *testing.T) {
	gm := setup()
	inBoundsCoordEdge := shared.Coord{0,50}
	response := gm.IsInBounds(inBoundsCoordEdge)
	if !response {
		t.Fail()
	}
	inBoundsCoordEdge = shared.Coord{99,25}
	response = gm.IsInBounds(inBoundsCoordEdge)
	if !response {
		t.Fail()
	}
}

func TestInBoundsEdgeY(t *testing.T) {
	gm := setup()
	inBoundsCoordEdge := shared.Coord{23,0}
	response := gm.IsInBounds(inBoundsCoordEdge)
	if !response {
		t.Fail()
	}
	inBoundsCoordEdge = shared.Coord{34,99}
	response = gm.IsInBounds(inBoundsCoordEdge)
	if !response {
		t.Fail()
	}
}

func TestOutOfBoundsX(t *testing.T) {
	gm := setup()
	outOfBoundsCoordEdge := shared.Coord{-1,45}
	response := gm.IsInBounds(outOfBoundsCoordEdge)
	if response {
		t.Fail()
	}
	outOfBoundsCoordEdge = shared.Coord{100,93}
	response = gm.IsInBounds(outOfBoundsCoordEdge)
	if response {
		t.Fail()
	}
}
func TestOutOfBoundsY(t *testing.T) {
	gm := setup()
	outOfBoundsCoordEdge := shared.Coord{68,-1}
	response := gm.IsInBounds(outOfBoundsCoordEdge)
	if response {
		t.Fail()
	}
	outOfBoundsCoordEdge = shared.Coord{34,100}
	response = gm.IsInBounds(outOfBoundsCoordEdge)
	if response {
		t.Fail()
	}
}

func TestWayOutOfBounds(t *testing.T) {
	gm := setup()
	wayOutOfBoundsCoord := shared.Coord{68, -12}
	response := gm.IsInBounds(wayOutOfBoundsCoord)
	if response {
		t.Fail()
	}
	wayOutOfBoundsCoord = shared.Coord{123, 39}
	response = gm.IsInBounds(wayOutOfBoundsCoord)
	if response {
		t.Fail()
	}
}

func TestNotWall(t *testing.T) {
	gm := setup()
	notWallCoord := shared.Coord{68, 43}
	response := gm.IsNotWall(notWallCoord)
	if !response {
		t.Fail()
	}
	notWallCoord = shared.Coord{12, 39}
	response = gm.IsNotWall(notWallCoord)
	if !response {
		t.Fail()
	}
	notWallCoord = shared.Coord{0, 11}
	response = gm.IsNotWall(notWallCoord)
	if !response {
		t.Fail()
	}
}

func TestIsWall(t *testing.T) {
	gm := setup()
	isWallCoord := shared.Coord{1, 1}
	response := gm.IsNotWall(isWallCoord)
	if response {
		t.Fail()
	}
	isWallCoord = shared.Coord{23, 99}
	response = gm.IsNotWall(isWallCoord)
	if response {
		t.Fail()
	}
}

func TestValidMove(t *testing.T) {
	gm := setup()
	validMove := shared.Coord{1, 2}
	response := gm.IsValidMove(validMove)
	if !response {
		t.Fail()
	}
	validMove = shared.Coord{25, 99}
	response = gm.IsValidMove(validMove)
	if !response {
		t.Fail()
	}
}

func TestInvalidMoveOutOfBounds(t *testing.T) {
	gm := setup()
	validMove := shared.Coord{1, -10}
	response := gm.IsValidMove(validMove)
	if response {
		t.Fail()
	}
	validMove = shared.Coord{100, 99}
	response = gm.IsValidMove(validMove)
	if response {
		t.Fail()
	}
}

func TestInvalidMoveWall(t *testing.T) {
	gm := setup()
	validMove := shared.Coord{1, 1}
	response := gm.IsValidMove(validMove)
	if response {
		t.Fail()
	}
	validMove = shared.Coord{10, 90}
	response = gm.IsValidMove(validMove)
	if response {
		t.Fail()
	}
}

func TestMoveTeleporting(t *testing.T) {
	gm := setup()
	prevCoord := shared.Coord{1, 1}
	newCoord := shared.Coord{2, 2}
	response := gm.IsNotTeleporting(prevCoord, newCoord)
	if response {
		fmt.Println("Nooo")
		t.Fail()
	}
	prevCoord = shared.Coord{1, 1}
	newCoord = shared.Coord{1, 3}
	response = gm.IsNotTeleporting(prevCoord, newCoord)
	if response {
		fmt.Println("POO")
		t.Fail()
	}
}

func TestIsNotTeleporting(t *testing.T) {
	gm := setup()
	prevCoord := shared.Coord{1, 1}
	newCoord := shared.Coord{1, 2}
	response := gm.IsNotTeleporting(prevCoord, newCoord)
	if !response {
		t.Fail()
	}
	prevCoord = shared.Coord{1, 1}
	newCoord = shared.Coord{2, 1}
	response = gm.IsNotTeleporting(prevCoord, newCoord)
	if !response {
		t.Fail()
	}
}

func TestGetNewPosition(t *testing.T) {
	gm := setup()
	steps := 3000 / 30 // number of steps on the game board
	i:=0
	// Try 10 new coordinates, make sure they are all valid
	for i < 10 {
		pos := gm.GetNewPos(shared.Coord{99,99})
		if !gm.IsValidMove(pos) {
			fmt.Println("Invalid position returned")

		}

		if pos.X < 1 + (steps/4) && pos.Y < 1 + (steps/4) {
			fmt.Println("New position not far enough away, fail")
			t.Fail()
		}

		i++
	}
}