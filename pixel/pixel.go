package main

import (
	"./impl"
	"os"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel"
	"fmt"
	"../shared"
	"golang.org/x/image/colornames"
	"image"
)

var nodeAddr string // must store as global to get it into run function

func main() {
	if len(os.Args) < 2 {
		nodeAddr = "127.0.0.1:12345" // use port 12345 on localhost for remote node if no input provided
	} else {
		nodeAddr = os.Args[1]
	}
	pixelgl.Run(run)
}

func run() {
	node := impl.CreatePixelNode(nodeAddr)
	go node.RunRemoteNodeListener()
	winMaxX := node.Geom.GetX()
	winMaxY := node.Geom.GetY()

	// all of our code will be fired up from here
	cfg := pixelgl.WindowConfig{
		Title:  "Wolfpack",
		Bounds: pixel.R(0, 0, winMaxX, winMaxY),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		fmt.Println(err)
	}
	win.Clear(colornames.Skyblue)

	//Enable text
	//basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	//basicTxt := text.New(geom.GetVectorFromCoords(1,1), basicAtlas)

	// Create player sprite
	pic, err := LoadPicture("../sprites/wolf.jpg")
	if err != nil {
		panic(err)
	}
	sprite := pixel.NewSprite(pic, pic.Bounds())

	node.PlayerSprite = sprite
	spritePos := node.Geom.GetVectorFromCoords(shared.Coord{3,3}) // starting position of sprite on grid

	// Create prey sprite
	pic, err = LoadPicture("../sprites/prey.jpg")
	if err != nil {
		panic(err)
	}
	preySprite := pixel.NewSprite(pic, pic.Bounds())
	node.PreySprite = preySprite
	preyPos := node.Geom.GetVectorFromCoords(shared.Coord{5,5})

	// Create other player sprite
	pic, err = LoadPicture("../sprites/other-player.jpg")
	if err != nil {
		panic(err)
	}
	otherPlayerSprite := pixel.NewSprite(pic, pic.Bounds())
	node.OtherPlayerSprite = otherPlayerSprite

	// Create wall sprite
	pic, err = LoadPicture("../sprites/wall.jpg")
	if err != nil {
		panic(err)
	}
	wallSprite := pixel.NewSprite(pic, pic.Bounds())
	node.WallSprite = wallSprite

	node.DrawWalls(win) // call this to draw walls every update

	sprite.Draw(win, pixel.IM.Moved(spritePos))
	preySprite.Draw(win, pixel.IM.Moved(preyPos))

	win.Update()

	for !win.Closed() {
		// Listens for keypress
		keyStroke := ""
		if win.Pressed(pixelgl.KeyLeft) {
			keyStroke = "left"
		} else if win.Pressed(pixelgl.KeyRight)  {
			keyStroke = "right"
		} else if win.Pressed(pixelgl.KeyUp) {
			keyStroke = "up"
		} else if win.Pressed(pixelgl.KeyDown) {
			keyStroke = "down"
		}
		if keyStroke != "" {
			node.SendMove(keyStroke)
			fmt.Println("sending keystroke")
			keyStroke = ""
		}
		// Update game state
		if len(node.NewGameStates) > 0 {
			curState := <- node.NewGameStates
			node.GameState = curState // set current state to the new state
			// Now, update the rendering
			node.RenderNewState(win)
		}
		node.RenderNewState(win)
		win.Update() // must be called frequently, or pixel will hang (can't update only when there is a new gamestate)
	}
}


func checkForWin(sprite pixel.Vec, prey pixel.Vec) (bool) {
	if sprite.X == prey.X && sprite.Y == prey.Y {
		return true
	}
	return false
}

func LoadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}