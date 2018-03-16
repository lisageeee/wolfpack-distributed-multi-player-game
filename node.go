package main

import (
	"fmt"
	"net"
	"os"
	"net/rpc"
	"strconv"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image"
	"./geometry"
	"./shared"
	"./prey"
	_ "image/png"
	_ "image/jpeg"
	"time"
	"log"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
)

var initialState shared.InitialState

func main() {
	fmt.Println("hello world")

	// Listener IP address
	var ip_addr string
	// Can start with an IP as param
	if (len(os.Args)>1){
		ip_addr = os.Args[1]
	}else{
		ip_addr = "127.0.0.1:0"
	}
	_, client := startListener(ip_addr)
	defer client.Close()

	go RunListener(client)

	gameConfig := ServerRegister(client.LocalAddr().String())
	otherNodes := gameConfig.Connections
	uniqueId := gameConfig.Identifier
	fmt.Println("Your identifier is:")
	fmt.Println(uniqueId)
	fmt.Println("The connections:")
	fmt.Println(otherNodes)
	initialState = gameConfig.InitState
	udpAddr := client.LocalAddr().(*net.UDPAddr)
	floodNodes(otherNodes, udpAddr)

	pixelgl.Run(run)
	select {}
}
func run() {

	// Window size
	var winMaxX float64 = initialState.Settings.WindowsX
	var winMaxY float64 = initialState.Settings.WindowsY

	// Sprite size
	var spriteMin float64 = 20
	var spriteMax float64 = 50
	spriteStep := spriteMax - spriteMin // winMaxX % spriteStep and winMaxY % spriteStep should be 0 (spriteStep == spriteSize)


	// Init walls
	wallCoords := initialState.Settings.WallCoordinates
	wallPic, err := loadPicture("./sprites/wall.jpg")
	if err != nil {
		panic(err)
	}

	// Create geometry manager
	geom := geometry.CreateGeometryManager(winMaxX, winMaxY, spriteStep, wallCoords)

	// Create walls sprites for drawing
	walls := createWallSprites(wallCoords, wallPic)
	wallVecs := geom.GetWallVectors()

	// all of our code will be fired up from here
	cfg := pixelgl.WindowConfig{
		Title:  "Wolfpack",
		Bounds: pixel.R(0, 0, winMaxX, winMaxY),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	win.Clear(colornames.Skyblue)

	//Enable text
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicTxt := text.New(geom.GetVectorFromCoords(1,1), basicAtlas)

	// Create player sprite
	pic, err := loadPicture("./sprites/bunny.jpeg")
	if err != nil {
		panic(err)
	}
	sprite := pixel.NewSprite(pic, pixel.R(spriteMin, spriteMin,spriteMax,spriteMax))
	spritePos := geom.GetVectorFromCoords(3,3) // starting position of sprite on grid
	if !geom.IsInBounds(spritePos) {
		spritePos = geom.GetVectorFromCoords(0, 0 ) // will always be in bounds if given incorrect args
	}

	// Create prey sprite
	pic, err = loadPicture("./sprites/prey.jpg")
	if err != nil {
		panic(err)
	}
	preySprite := pixel.NewSprite(pic, pic.Bounds())
	preyRunner := prey.CreatePreyRunner(geom)
	preySprite.Draw(win, pixel.IM.Moved(preyRunner.GetPosition()))

	drawWalls(wallVecs, walls, win) // call this to draw walls every update
	sprite.Draw(win, pixel.IM.Moved(spritePos))


	keyStroke := ""
	ttm := make(chan string, 1)

	for !win.Closed() {


		go func() {
			time.Sleep(150*time.Millisecond)
			ttm <- "ok"
		}()

		// Listens for keypress
		go func() {
			if win.Pressed(pixelgl.KeyLeft) {
				keyStroke = "left"
			}
			if win.Pressed(pixelgl.KeyRight)  {
				keyStroke = "right"
			}
			if win.Pressed(pixelgl.KeyUp) {
				keyStroke = "up"
			}
			if win.Pressed(pixelgl.KeyDown) {
				keyStroke = "down"
			}
		}()

		select {
		case _ = <-ttm:
			win.Clear(colornames.Skyblue)
			mat := pixel.IM
			newLoc := spritePos
			switch keyStroke {
				case "up":
					newLoc = pixel.V(spritePos.X, spritePos.Y + spriteStep)
					if geom.IsInBounds(newLoc) && !geom.IsCollision(newLoc) {
						spritePos.Y = spritePos.Y + spriteStep
					}
				case "down":
					newLoc = pixel.V(spritePos.X, spritePos.Y - spriteStep)
					if geom.IsInBounds(newLoc) && !geom.IsCollision(newLoc) {
						spritePos.Y = spritePos.Y - spriteStep
					}
				case "left":
					newLoc = pixel.V(spritePos.X - spriteStep, spritePos.Y)
					if geom.IsInBounds(newLoc) && !geom.IsCollision(newLoc) {
						spritePos.X = spritePos.X - spriteStep
					}
				case "right":
					newLoc = pixel.V(spritePos.X + spriteStep, spritePos.Y)
					if geom.IsInBounds(newLoc) && !geom.IsCollision(newLoc) {
						spritePos.X = spritePos.X + spriteStep
					}
			}
			keyStroke = ""
			mat = mat.Moved(spritePos)
			preyPos := preyRunner.Move()
			preySprite.Draw(win, pixel.IM.Moved(preyPos))
			drawWalls(wallVecs, walls, win)
			sprite.Draw(win, mat)

			// Check for win condition
			if checkForWin(spritePos, preyPos) {
				fmt.Fprintln(basicTxt, "You win!")
				basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 4))
				win.Update()
				preyRunner = prey.CreatePreyRunner(geom)
				time.Sleep(2*time.Second)
				basicTxt.Clear()
			}
			win.Update()
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

func floodNodes(otherNodes []string, udp_addr *net.UDPAddr) {
	localIP, _ := net.ResolveUDPAddr("udp", udp_generic)
	for _, ip := range otherNodes {
		node_udp, _ := net.ResolveUDPAddr("udp", ip)
		// Connect to other node
		node_client, err := net.DialUDP("udp", localIP, node_udp)
		if err != nil {
			panic(err)
		}
		// Exchange messages with other node
		myListener := udp_addr.IP.String() + ":" +  strconv.Itoa(udp_addr.Port)
		node_client.Write([]byte(myListener))
	}
}

func ServerRegister(localIP string) shared.GameConfig {
	// Connect to server with RPC, port is always :8081
	serverConn, err := rpc.Dial("tcp", ":8081")
	if err != nil {
		log.Println("Cannot dial server. Please ensure the server is running and try again.")
		os.Exit(1)
	}
	var response shared.GameConfig
	// Get IP from server
	err = serverConn.Call("GServer.Register", localIP, &response)
	if err != nil {
		panic(err)
	}

	return response
}

func startListener(ip_addr string) (*net.UDPAddr, *net.UDPConn) {
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

const udp_generic = "127.0.0.1:0"
var clients []*net.Conn
func RunListener(client *net.UDPConn) {
	// takes a listener client
	// runs the listener in a infinite loop

	client.SetReadBuffer(1048576)

	i := 0
	for {
		i++
		buf := make([]byte, 1024)
		rlen, addr, err := client.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(buf[0:rlen]))
		fmt.Println(addr)
		fmt.Println(i)
		if string(buf[0:rlen]) != "connected" {
			remote_client, err := net.Dial("udp", string(buf[0:rlen]))
			if err != nil {
				panic(err)
			}
			remote_client.Write([]byte("connected"))

			clients = append(clients, &remote_client)
		}
	}
}

