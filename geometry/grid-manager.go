package geometry

import (
	"../shared"
	"strconv"
)

type GridManager struct {
	x int
	y int
	walls map[string]shared.Coord
}

const playerSize int = 30

// x int, y int, walls []shared.Coord
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

func (gm * GridManager) IsInBounds(coord shared.Coord) (bool) {
	if coord.X >= 0 && coord.X < gm.x {
		if coord.Y >= 0 && coord.Y < gm.y {
			return true
		}
	}
	return false
}

func (gm * GridManager) IsNotWall(coord shared.Coord) (bool) {
	// Convert coord to string, check map
	key := strconv.Itoa(coord.X) + " " + strconv.Itoa(coord.Y)

	// Check map for coord
	_, ok := gm.walls[key]

	return !ok
}

func (gm * GridManager) IsNotTeleporting(origCoord shared.Coord, newCoord shared.Coord) (bool) {
	x := origCoord.X - newCoord.X
	y := origCoord.Y - newCoord.Y
	if ((x+y) < -1 || (x+y) > 1) && x != 0 && y != 0 {
		return false
	} else {
		return true
	}
}

func (gm * GridManager) IsValidMove(coord shared.Coord) (bool) {
	return gm.IsInBounds(coord) && gm.IsNotWall(coord)
}