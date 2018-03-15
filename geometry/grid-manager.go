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

func CreateNewGridManager(x int, y int, walls []shared.Coord) (GridManager) {
	wallMap := make(map[string]shared.Coord)

	// Create the map of walls for fast lookup
	for _, wall := range(walls) {
		key := strconv.Itoa(wall.X) + " " + strconv.Itoa(wall.Y)
		wallMap[key] = wall
	}

	gm := GridManager{x: x, y: y, walls: wallMap}
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

func (gm * GridManager) IsValidMove(coord shared.Coord) (bool) {
	return gm.IsInBounds(coord) && gm.IsNotWall(coord)
}