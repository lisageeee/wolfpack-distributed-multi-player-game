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

// Sprite size
const spriteStep = 30


type PixelNode struct {
	//Listener          *net.UDPConn
	Sender            *net.TCPConn
	playerPosition    shared.Coord
	Geom              geometry.PixelManager
	GameState         shared.GameRenderState
	NewGameStates     chan shared.GameRenderState
	PlayerSprite      *pixel.Sprite
	WallSprite        *pixel.Sprite
	OtherPlayerSprite *pixel.Sprite
	PreySprite        *pixel.Sprite
}

func CreatePixelNode(nodeAddr string) (PixelNode) {

	// Setup connection
	remote := setupTCP(nodeAddr)

	// Get initial game state (first message after tcp setup)
	var buf = make([]byte, 2048)
	var settings shared.InitialGameSettings

	// Get initial game state
	rlen, err := remote.Read(buf)
	if err != nil {
		fmt.Println(err)
	} else {
		json.Unmarshal(buf[0:rlen], &settings)
	}

	// Init walls
	wallCoords := settings.WallCoordinates

	// Create geometry manager
	geom := geometry.CreatePixelManager(settings.WindowsX, settings.WindowsY, spriteStep, wallCoords)

	node := PixelNode{ Sender: remote, Geom: geom, NewGameStates: make(chan shared.GameRenderState, 5)}

	// go node.RunRemoteNodeListener()
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
func (pn * PixelNode) RunRemoteNodeListener() {
	// takes a Listener client
	// runs the Listener in a infinite loop
	node := pn.Sender

	i := 0
	var playerPos shared.GameRenderState
	for {
		i++
		buf := make([]byte, 1024)
		rlen, err := node.Read(buf)
		fmt.Println(node)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(buf[0:rlen], &playerPos)
		if err != nil {
			fmt.Println("Error receiving new GameRenderState on PixelNode:", err)

		} else {
			pn.NewGameStates <- playerPos
		}
	}
}

func (pn * PixelNode ) DrawWalls(window *pixelgl.Window) {
	for _, wall := range pn.Geom.GetWallVectors() {
		pn.WallSprite.Draw(window, pixel.IM.Moved(wall))
	}
}

func setupTCP(ip_addr string) (*net.TCPConn) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip_addr)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(tcpAddr)
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return tcpConn
}