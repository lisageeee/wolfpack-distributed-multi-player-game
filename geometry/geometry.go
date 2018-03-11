package geometry

import "github.com/faiface/pixel"

type GeometryManager struct {
	x float64
	y float64
	spriteSize float64
}

func CreateGeometryManager(x float64, y float64, spriteSize float64) (GeometryManager) {
	gm := GeometryManager{x: x, y: y, spriteSize: spriteSize}
	return gm
}

func (gm * GeometryManager) GetVectorFromCoords(x float64, y float64) (pixel.Vec) {
	xVec := x * gm.spriteSize + 0.5 * gm.spriteSize
	yVec := y * gm.spriteSize + 0.5 * gm.spriteSize
	vec := pixel.V(xVec, yVec)
	return vec
}