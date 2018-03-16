package geometry

import (
	"github.com/faiface/pixel"
	"../shared"
)

type PixelManager struct {
	x float64
	y float64
	spriteSize float64
	wallVectors []pixel.Vec
}

// Creates a new instance of a Geometry manager to handle movement.
// Takes the Pixel max dimensions of the window, the one-dimensional size of the square sprite and an array of wall coords.
func CreatePixelManager(windowMaxX float64, windowMaxY float64, spriteSize float64, walls []shared.Coord) (PixelManager) {
	gm := PixelManager{x: windowMaxX, y: windowMaxY, spriteSize: spriteSize}
	gm.getWallVectors(walls)
	return gm
}

// Takes a coordinate (x, y) value from the 'Wolfpack' grid and returns a Pixel-understandable vector.
func (gm *PixelManager) GetVectorFromCoords(coord shared.Coord) (pixel.Vec) {
	xVec := float64(coord.X) * gm.spriteSize + 0.5 * gm.spriteSize
	yVec := float64(coord.Y) * gm.spriteSize + 0.5 * gm.spriteSize
	vec := pixel.V(xVec, yVec)
	return vec
}

// Takes a shared.Coord array and converts to a pixel vector array.
// Assigns the vector array to the local "walls" attr.
func (gm *PixelManager) getWallVectors(walls []shared.Coord) {
	wallVecs := make([]pixel.Vec, len(walls))
	for i, wall := range walls {
		vec := gm.GetVectorFromCoords(wall)
		wallVecs[i] = vec
	}
	gm.wallVectors = wallVecs
}

func (gm *PixelManager) GetWallVectors() ([]pixel.Vec) {
	return gm.wallVectors
}

func (gm *PixelManager) GetX() (float64) {
	return gm.x
}

func (gm *PixelManager) GetY() (float64) {
	return gm.y
}

func (gm *PixelManager) GetSpriteSize() (float64)  {
	return gm.spriteSize
}