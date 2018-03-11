package geometry

import (
	"github.com/faiface/pixel"
	"../shared"
	)


type GeometryManager struct {
	x float64
	y float64
	spriteSize float64
	walls []pixel.Vec
}

// Creates a new instance of a Geometry manager to handle movement.
// Takes the Pixel max dimensions of the window, the one-dimensional size of the square sprite and an array of wall coords.
func CreateGeometryManager(windowMaxX float64, windowMaxY float64, spriteSize float64, walls []shared.Coord) (GeometryManager) {
	gm := GeometryManager{x: windowMaxX, y: windowMaxY, spriteSize: spriteSize}
	gm.getWallVectors(walls)
	return gm
}

// Takes a coordinate (x, y) value from the 'Wolfpack' grid and returns a Pixel-understandable vector.
func (gm * GeometryManager) GetVectorFromCoords(x float64, y float64) (pixel.Vec) {
	xVec := x * gm.spriteSize + 0.5 * gm.spriteSize
	yVec := y * gm.spriteSize + 0.5 * gm.spriteSize
	vec := pixel.V(xVec, yVec)
	return vec
}

// Takes a Pixel vector and determines whether or not the sprite will be completely in bounds at this location.
// Returns true if the sprite would be in bounds, false otherwise.
func (gm * GeometryManager) IsInBounds(loc pixel.Vec) (bool) {
	if loc.Y < gm.spriteSize/2 || loc.X < gm.spriteSize / 2 {
		return false
	} else if loc.Y > gm.x - gm.spriteSize /2 || loc.X > gm.x - gm.spriteSize / 2 {
		return false
	}
	return true
}

// Checks if a given move creates a collision with a wall.
// Returns true if a wall collision has happened, false otherwise
func (gm * GeometryManager) IsCollision(loc pixel.Vec) (bool) {
	for _, wall := range gm.walls {
		if wall.Y == loc.Y && wall.X == loc.X {
			return true
		}
	}
	return false
}

// Takes a shared.Coord array and converts to a pixel vector array.
// Assigns the vector array to the local "walls" attr.
func (gm * GeometryManager) getWallVectors(walls []shared.Coord) {
	wallVecs := make([]pixel.Vec, len(walls))
	for i, wall := range walls {
		vec := gm.GetVectorFromCoords(wall.X, wall.Y)
		wallVecs[i] = vec
	}
	gm.walls = wallVecs
}

func (gm * GeometryManager) GetWallVectors() ([]pixel.Vec) {
	return gm.walls
}

func (gm * GeometryManager) GetX() (float64) {
	return gm.x
}

func (gm * GeometryManager) GetY() (float64) {
	return gm.y
}

func (gm * GeometryManager) GetSpriteSize() (float64)  {
	return gm.spriteSize
}