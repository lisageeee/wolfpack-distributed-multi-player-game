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
	scoreboardWidth float64
}

// Creates a new instance of a Geometry manager to handle movement.
// Takes the Pixel max dimensions of the window, the one-dimensional size of the square sprite and an array of wall coords.
func CreatePixelManager(windowMaxX, windowMaxY, scoreboardWidth float64, spriteSize float64, walls []shared.Coord) (PixelManager) {
	gm := PixelManager{x: windowMaxX, y: windowMaxY, spriteSize: spriteSize, scoreboardWidth: scoreboardWidth}
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

// Returns all wall vectors for rendering.
func (gm *PixelManager) GetWallVectors() ([]pixel.Vec) {
	return gm.wallVectors
}

// Returns the maximum gamespace X value (in pixels) as a float64
func (gm *PixelManager) GetX() (float64) {
	return gm.x
}

// Returns the maximum gamespace Y value (in pixels) as a float64
func (gm *PixelManager) GetY() (float64) {
	return gm.y
}

// Returns the width of the scoreboard (in pixels) as a float64
func (gm *PixelManager) GetScoreboardWidth() (float64) {
	return gm.scoreboardWidth
}

// Returns the width/height of a sprite, and thus the grid the game is played on, as a float64
func (gm *PixelManager) GetSpriteSize() (float64)  {
	return gm.spriteSize
}