package main

import (
	"os"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image"
	"./geometry"
	"./shared"
	_ "image/png"
	_ "image/jpeg"
	"net"
	"fmt"
	"encoding/json"
)

var nodeAddr string // must store as global to get it into run function
var myAddr string

func main() {
	if len(os.Args) < 3 {
		nodeAddr = "127.0.0.1:12345" // use port 12345 on localhost for remote node if no input provided
		myAddr = "127.0.0.1:1234" // use :1234 for incoming messages
	} else {
		nodeAddr = os.Args[1]
	}
	pixelgl.Run(run)
}

type PixelNode struct {
	listener *net.UDPConn
	sender *net.UDPConn
	playerPosition shared.Coord
	geom geometry.PixelManager
	GameState shared.GameRenderState
	newGameStates chan shared.GameRenderState
	PlayerSprite *pixel.Sprite
}

func run() {
	// Window size
	var winMaxX float64 = 300
	var winMaxY float64 = 300

	// Sprite size
	var spriteMin float64 = 20
	var spriteMax float64 = 50

	spriteStep := spriteMax - spriteMin // winMaxX % spriteStep and winMaxY % spriteStep should be 0 (spriteStep == spriteSize)


	// Init walls
	wallCoords := []shared.Coord{{X: 4, Y:3}}
	wallPic, err := loadPicture("./sprites/wall.jpg")
	if err != nil {
		panic(err)
	}

	// Create geometry manager
	geom := geometry.CreatePixelManager(winMaxX, winMaxY, spriteStep, wallCoords)
	//
	_, conn := startListen(myAddr)
	remote := setupUDP(nodeAddr)
	node := PixelNode{listener:conn, sender: remote, geom: geom, newGameStates: make(chan shared.GameRenderState, 5)}
	go node.runRemoteNodeListener()

	// Create walls sprites for drawing
	walls := createWallSprites(wallCoords, wallPic)
	wallVecs := geom.GetWallVectors()

	// all of our code will be fired up from here
	cfg := pixelgl.WindowConfig{
		Title:  "Wolfpack",
		Bounds: pixel.R(0, 0, 300, 300),
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
	pic, err := loadPicture("./sprites/bunny.jpeg")
	if err != nil {
		panic(err)
	}
	sprite := pixel.NewSprite(pic, pixel.R(spriteMin, spriteMin,spriteMax,spriteMax))

	node.PlayerSprite = sprite
	spritePos := geom.GetVectorFromCoords(shared.Coord{3,3}) // starting position of sprite on grid

	// Create prey sprite
	pic, err = loadPicture("./sprites/prey.jpg")
	if err != nil {
		panic(err)
	}

	drawWalls(wallVecs, walls, win) // call this to draw walls every update

	sprite.Draw(win, pixel.IM.Moved(spritePos))

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
			node.sender.Write([]byte(keyStroke))
			fmt.Println("sending keystroke")
			keyStroke = ""
		}

		// Update game state
		if len(node.newGameStates) > 0 {
			curState := <- node.newGameStates
			node.GameState = curState
			win.Clear(colornames.Skyblue)

			drawWalls(wallVecs, walls, win)
			node.renderNewState(win)
		}
		win.Update() // must be called frequently, or pixel will hang (can't update only when there is a new gamestate)
	}
}

func (pn * PixelNode) renderNewState(win * pixelgl.Window) {
	curState := pn.GameState
	// Render walls, first
	playerPos := pn.geom.GetVectorFromCoords(curState.PlayerLoc)
	mat := pixel.IM
	mat = mat.Moved(playerPos)
	pn.PlayerSprite.Draw(win, mat)
}


// Listens for new game states from pixel node
func (pn * PixelNode) runRemoteNodeListener() {
	// takes a listener client
	// runs the listener in a infinite loop
	node := pn.listener
	node.SetReadBuffer(1048576)

	i := 0
	var playerPos shared.GameRenderState
	for {
		i++
		buf := make([]byte, 1024)
		rlen, _, err := node.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(buf[0:rlen], &playerPos)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Error w/new coord")
		} else {
			fmt.Println("Got new coord")
			pn.newGameStates <- playerPos
		}
	}
}

func checkForWin(sprite pixel.Vec, prey pixel.Vec) (bool) {
	if sprite.X == prey.X && sprite.Y == prey.Y {
		return true
	}
	return false
}

// Creates an array of sprites to be used for the walls
func createWallSprites(coords []shared.Coord, picture pixel.Picture) ([]pixel.Sprite) {
	sprites := make([]pixel.Sprite, len(coords))
	for i, _ := range coords {
		sprites[i] = *pixel.NewSprite(picture, picture.Bounds())
	}
	return sprites
}

func drawWalls(vectors []pixel.Vec, sprites []pixel.Sprite, window *pixelgl.Window) {
	for i := range vectors {
		sprites[i].Draw(window, pixel.IM.Moved(vectors[i]))
	}
}

func loadPicture(path string) (pixel.Picture, error) {
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


func startListen(ip_addr string) (*net.UDPAddr, *net.UDPConn) {
	// takes an ip address and port to listen on
	// returns the udp address and listener client
	// starts Listener
	udp_addr, _ := net.ResolveUDPAddr("udp", ip_addr)
	client, err := net.ListenUDP("udp", udp_addr)
	if err != nil {
		panic(err)
	}
	return udp_addr, client
}

func setupUDP(ip_addr string) (*net.UDPConn) {
	node_udp, _ := net.ResolveUDPAddr("udp", ip_addr)
	node_client, err := net.DialUDP("udp", nil, node_udp)
	if err != nil {
		panic(err)
	}
	return node_client
}