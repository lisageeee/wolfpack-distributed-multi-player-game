package prey

import (
	"../geometry"
	"github.com/faiface/pixel"
)

type PreyRunner struct {
	geo geometry.PixelManager
	position pixel.Vec
}

func (pr *PreyRunner) GetPosition() (pixel.Vec) {
	return pr.position
}

func (pr *PreyRunner) Move() (pixel.Vec) {
	//random := rand.Float64()
	//X := pr.position.X
	//Y:= pr.position.Y
	//step := pr.geo.GetSpriteSize()
	//var newVec pixel.Vec
	//switch {
	//	case random < 0.25:
	//		newVec = pixel.V(X - step, Y + step)
	//	case random < 0.5:
	//		newVec = pixel.V(X + step, Y + step)
	//	case random < 0.75:
	//		newVec = pixel.V(X + step, Y - step)
	//	default:
	//		newVec = pixel.V(X - step, Y - step)
	//}
	return pr.position
}