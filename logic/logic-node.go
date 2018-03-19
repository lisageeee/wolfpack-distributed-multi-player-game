package main

import (
	"fmt"
	"net"
	"os"
	"net/rpc"
	"strconv"
	_ "image/png"
	_ "image/jpeg"
	"log"
	"../shared"
	"../geometry"
	l "./impl"
	"encoding/json"
)

type RemotePlayerInterface struct {
	pixelInterface	  l.PixelInterface
	playerCommChannel chan string
	GameRenderState	  shared.GameRenderState
	geo               geometry.GridManager
	identifier        int
	GameConfig		  shared.InitialState
}

// Entrypoint, sets up communication channels and creates the RemotePlayerInterface
func main() {
	fmt.Println("hello world")

	// Listener IP address
	var node_listener_ip_address string
	var player_listener_ip_address string
	var pixel_ip_address string
	// Can start with an IP as param
	if len(os.Args) > 2 {
		node_listener_ip_address = os.Args[1]
		player_listener_ip_address = os.Args[2]
	} else if len(os.Args)>1{
		node_listener_ip_address = "127.0.0.1:0"
		player_listener_ip_address = os.Args[1]
		pixel_ip_address = "127.0.0.1:1234"
	} else {
		node_listener_ip_address = "127.0.0.1:0"
		player_listener_ip_address = "127.0.0.1:12345"
		pixel_ip_address = "127.0.0.1:1234"
	}

	// Setup the player communcation channel
	playerCommChannel := make(chan string)

	// Startup Pixel interface + listening
	pixelInterface := l.CreatePixelInterface(playerCommChannel)
	go pixelInterface.RunPlayerListener(pixel_ip_address, player_listener_ip_address)

	// Start the node to node interface
	nodeInterface := l.CreateNodeCommInterface()
	go nodeInterface.RunListener(node_listener_ip_address)

	gameConfig := nodeInterface.ServerRegister()
	uniqueId := gameConfig.Identifier
	initState := gameConfig.InitState

	// Make default gameState
	gameRenderState := shared.GameRenderState{PlayerLoc:shared.Coord{3,3},
	OtherPlayers: []shared.Coord{{6,6}}, Prey: shared.Coord{5,5}} // TODO: change these to dynamic when
																				// we connect to other players/prey
	pi := RemotePlayerInterface{pixelInterface: pixelInterface,
	playerCommChannel: playerCommChannel, geo: geometry.CreateNewGridManager(initState.Settings),
	GameRenderState: gameRenderState, identifier: uniqueId, GameConfig: initState}
	pi.runGame()
}

func (pi * RemotePlayerInterface) runGame() {

	for {
		message := <-pi.playerCommChannel
		switch message {
		case "quit":
			break
		default:
			pi.movePlayer(message)
			pi.pixelInterface.SendPlayerGameState(pi.GameRenderState)
			fmt.Println("movin' player", message)
		}
	}

}

func (pi * RemotePlayerInterface) movePlayer(move string) {
	// Get current player state
	playerLoc := pi.GameRenderState.PlayerLoc

	// Calculate new position with move
	newPosition := shared.Coord{X: playerLoc.X, Y: playerLoc.Y}
	fmt.Println(move)
	switch move {
		case "up":
			newPosition.Y = newPosition.Y + 1
		case "down":
			newPosition.Y = newPosition.Y - 1
		case "left":
			newPosition.X = newPosition.X - 1
		case "right":
			newPosition.X = newPosition.X + 1
	}
	// Check new move is valid, if so update player position
	if pi.geo.IsValidMove(newPosition) {
		pi.GameRenderState.PlayerLoc = newPosition
	}
}