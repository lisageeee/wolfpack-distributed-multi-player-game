package geometry

import (
	"../shared"
	"strconv"
	"math/rand"
	"math"
)

type GridManager struct {
	x int
	y int
	walls map[string]shared.Coord
}

const playerSize int = 30

// Creates a new grid manager for use in a logic node. Can perform checks on proposed coordinates.
// Returns the created grid manager
func CreateNewGridManager(settings shared.InitialGameSettings) (GridManager) {
	wallMap := make(map[string]shared.Coord)

	// Create the map of walls for fast lookup
	for _, wall := range(settings.WallCoordinates) {
		key := strconv.Itoa(wall.X) + " " + strconv.Itoa(wall.Y)
		wallMap[key] = wall
	}

	// Figure out how big our grid is
	gridX := int(settings.WindowsX) / playerSize
	gridY := int(settings.WindowsY) / playerSize

	gm := GridManager{x: gridX, y: gridY, walls: wallMap}
	return gm
}

// Checks if a given coordinate is in bounds on the current game board
// Returns true if the coordinate is in bounds, false otherwise.
func (gm * GridManager) IsInBounds(coord shared.Coord) (bool) {
	if coord.X >= 0 && coord.X < gm.x {
		if coord.Y >= 0 && coord.Y < gm.y {
			return true
		}
	}
	return false
}

// Checks if a given coordinate not the same as a wall coordinate on the current map
// Returns true if the coordinate is not the same as a wall coordinate, false otherwise
func (gm * GridManager) IsNotWall(coord shared.Coord) (bool) {
	// Convert coord to string, check map
	key := strconv.Itoa(coord.X) + " " + strconv.Itoa(coord.Y)

	// Check map for coord
	_, ok := gm.walls[key]

	return !ok
}

// Checks that the two given coordinates could be valid new and original states; that is, ensures
// the player isn't taking more than one step per move
// Return true if the move is valid, false if the node has been "teleporting"
func (gm * GridManager) IsNotTeleporting(origCoord shared.Coord, newCoord shared.Coord) (bool) {
	x := origCoord.X - newCoord.X
	y := origCoord.Y - newCoord.Y
	if x < -1 || y < -1 || x > 1 || y > 1 {
		return false
	}
	if ((x+y) < -1 || (x+y) > 1) && x != 0 && y != 0 {
		return false
	} else {
		return true
	}
}

func (gm * GridManager) GetRandomValidPos() (shared.Coord) {
	var posX int
	var posY int
	for {
		posX = rand.Intn(gm.x)
		posY = rand.Intn(gm.y)
		if gm.IsValidMove(shared.Coord{posX, posY}) {
			break
		}
	}
	return shared.Coord{X:posX, Y:posY}
}


func (gm * GridManager) GetNewPos(oldMove shared.Coord) (shared.Coord) {
	posX := gm.x - oldMove.X
	posY := gm.y - oldMove.Y

	if gm.IsValidMove(shared.Coord{posX, posY}) && gm.isFarAway(oldMove.X, oldMove.Y, posX, posY) {
		return shared.Coord{X:posX, Y:posY}
	}

	var pos shared.Coord
	for {
		pos = gm.GetRandomValidPos()

		if gm.isFarAway(oldMove.X, oldMove.Y, pos.X, pos.Y) {
			break
		}
	}
	return pos
}

// Checks that a given move is valid by checking if it is in bounds and also not a wall
// Returns true of the move is valid, false otherwise.
func (gm * GridManager) IsValidMove(coord shared.Coord) (bool) {
	return gm.IsInBounds(coord) && gm.IsNotWall(coord)
}

// Returns true if the new point is at least a quarter of the playscreen away
func (gm* GridManager) isFarAway(oldX, oldY, newX, newY int) (bool) {
	if math.Abs(float64(newX-oldX)) > float64(gm.x/4) ||  math.Abs(float64(newY-oldY)) > float64(gm.y/4) {
		return true
	}
	return false
}