package geometry

import "github.com/faiface/pixel"

type GeometryManager struct {
	x float64
	y float64
	spriteSize float64
}

// Creates a new instance of a Geometry manager to handle movement.
// Takes the Pixel max dimensions of the window and the one-dimensional size of the square sprite.
func CreateGeometryManager(windowMaxX float64, windowMaxY float64, spriteSize float64) (GeometryManager) {
	gm := GeometryManager{x: windowMaxX, y: windowMaxY, spriteSize: spriteSize}
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