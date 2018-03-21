package impl

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"../../geometry"
	"../../shared"
	_ "image/png"
	_ "image/jpeg"
	"net"
	"fmt"
	"encoding/json"
)

var NodeAddr string // must store as global to get it into run function
var MyAddr string
// Window size

var WinMaxX float64 = 300
var WinMaxY float64 = 300

// Sprite size
var SpriteMin float64 = 20
var SpriteMax float64 = 50


type PixelNode struct {
	Listener          *net.UDPConn
	Sender            *net.UDPConn
	playerPosition    shared.Coord
	Geom              geometry.PixelManager
	GameState         shared.GameRenderState
	NewGameStates     chan shared.GameRenderState
	PlayerSprite      *pixel.Sprite
	WallSprite        *pixel.Sprite
	OtherPlayerSprite *pixel.Sprite
	PreySprite        *pixel.Sprite
}

func CreatePixelNode(nodeAddr string, myAddr string) (PixelNode) {
	spriteStep := SpriteMax - SpriteMin // WinMaxX % spriteStep and WinMaxY % spriteStep should be 0 (spriteStep == spriteSize)

	// Init walls
	wallCoords := []shared.Coord{{X: 4, Y:3}}

	// Create geometry manager
	geom := geometry.CreatePixelManager(WinMaxX, WinMaxY, spriteStep, wallCoords)
	//
	_, conn := startListen(myAddr)
	remote := setupUDP(nodeAddr)
	node := PixelNode{Listener:conn, Sender: remote, Geom: geom, NewGameStates: make(chan shared.GameRenderState, 5)}
	go node.runRemoteNodeListener()
	return node
}

//
func (pn * PixelNode) RenderNewState (win * pixelgl.Window) {
	curState := pn.GameState

	// Clear current render
	win.Clear(colornames.Skyblue)

	// Render walls, first
	pn.DrawWalls(win)

	// Render prey
	pn.PreySprite.Draw(win, pixel.IM.Moved(pn.Geom.GetVectorFromCoords(curState.Prey)))

	// Render other players
	for _, player := range curState.OtherPlayers {
		pn.OtherPlayerSprite.Draw(win, pixel.IM.Moved(pn.Geom.GetVectorFromCoords(player)))
	}

	// Render player
	playerPos := pn.Geom.GetVectorFromCoords(curState.PlayerLoc)
	mat := pixel.IM
	mat = mat.Moved(playerPos)
	pn.PlayerSprite.Draw(win, mat)

}

func (pn * PixelNode) SendMove (move string) {
	pn.Sender.Write([]byte(move))
}


// Listens for new game states from pixel node
func (pn * PixelNode) runRemoteNodeListener() {
	// takes a Listener client
	// runs the Listener in a infinite loop
	node := pn.Listener
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
			pn.NewGameStates <- playerPos
		}
	}
}

func (pn * PixelNode ) DrawWalls(window *pixelgl.Window) {
	for _, wall := range pn.Geom.GetWallVectors() {
		pn.WallSprite.Draw(window, pixel.IM.Moved(wall))
	}
}

func startListen(ip_addr string) (*net.UDPAddr, *net.UDPConn) {
	// takes an ip address and port to listen on
	// returns the udp address and Listener client
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