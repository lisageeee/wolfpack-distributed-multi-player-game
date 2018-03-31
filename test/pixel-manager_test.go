package test

import (
	"testing"
	"../geometry"
	"../shared"
	)

const sizeX  = 300
const sizeY = 300
const spriteSize = 30
const scoreboardWidth = 200

func pmTestSetup() (geometry.PixelManager) {
	// Creates pixel manager on a 300x300 screen with a 30px
	pm := geometry.CreatePixelManager(sizeX, sizeY, scoreboardWidth, spriteSize,
		[]shared.Coord{{1,1}, {10, 90}, {23, 99} })
	return pm
}

func TestGetX(t *testing.T) {
	pm := pmTestSetup()
	x := pm.GetX()
	if x != sizeX {
		t.Fail()
	}
}

func TestGetY(t *testing.T) {
	pm := pmTestSetup()
	y := pm.GetY()
	if y != sizeY {
		t.Fail()
	}
}


func TestGetLocationOrigin(t *testing.T) {
	pm := pmTestSetup()
	xLoc := 0
	yLoc := 0
	vec := pm.GetVectorFromCoords(shared.Coord{xLoc,yLoc})
	if vec.X != float64(xLoc) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
	if vec.Y != float64(yLoc) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
}

func TestGetLocationOther(t *testing.T) {
	pm := pmTestSetup()
	xLoc := 4
	yLoc := 2
	vec := pm.GetVectorFromCoords(shared.Coord{xLoc,yLoc})
	if vec.X != float64(xLoc) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
	if vec.Y != float64(yLoc) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
}

func TestWallVectors(t *testing.T) {
	pm := pmTestSetup()
	vecs := pm.GetWallVectors()
	if len(vecs) != 3 {
		t.FailNow()
	}
	if vecs[0].X != float64(1) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
	if vecs[0].Y != float64(1) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
	if vecs[1].X != float64(10) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
	if vecs[1].Y != float64(90) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
	if vecs[2].X != float64(23) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
	if vecs[2].Y != float64(99) * spriteSize + 0.5 * spriteSize {
		t.Fail()
	}
}
