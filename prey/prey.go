package prey

import (
	"../geometry"
	"github.com/faiface/pixel"
	"math/rand"
)

type PreyRunner struct {
	geo geometry.GeometryManager
	position pixel.Vec
}

func CreatePreyRunner(manager geometry.GeometryManager) (PreyRunner){
	pos := manager.GetRandomValidPosition()
	pr := PreyRunner{geo: manager, position: pos}
	return pr
}

func (pr *PreyRunner) GetPosition() (pixel.Vec) {
	return pr.position
}

func (pr *PreyRunner) Move() (pixel.Vec) {
	random := rand.Float64()
	X := pr.position.X
	Y:= pr.position.Y
	step := pr.geo.GetSpriteSize()
	var newVec pixel.Vec
	switch {
		case random < 0.25:
			newVec = pixel.V(X - step, Y + step)
		case random < 0.5:
			newVec = pixel.V(X + step, Y + step)
		case random < 0.75:
			newVec = pixel.V(X + step, Y - step)
		default:
			newVec = pixel.V(X - step, Y - step)
	}

	if pr.geo.IsInBounds(newVec) && !pr.geo.IsCollision(newVec) {
		pr.position = newVec
		return pr.position
	} else {
		return pr.Move()
	}
}