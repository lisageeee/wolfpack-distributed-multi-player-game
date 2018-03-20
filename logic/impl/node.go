package impl

import (
	"../../shared"
	"../../geometry"
	"fmt"
)

type PlayerNode struct {
	pixelInterface	  PixelInterface
	nodeInterface 	  NodeCommInterface
	playerCommChannel chan string
	GameRenderState	  shared.GameRenderState
	geo               geometry.GridManager
	identifier        int
	GameConfig		  shared.InitialState
}


func CreatePlayerNode(nodeListenerAddr, playerListenerAddr, pixelSendAddr string) (PlayerNode) {
	// Setup the player communication buffered channel
	playerCommChannel := make(chan string, 5)

	// Startup Pixel interface + listening
	pixelInterface := CreatePixelInterface(playerCommChannel)
	go pixelInterface.RunPlayerListener(pixelSendAddr, playerListenerAddr)

	// Start the node to node interface
	nodeInterface := CreateNodeCommInterface()
	go nodeInterface.RunListener(nodeListenerAddr)

	// Register with server, update info
	gameConfig := nodeInterface.ServerRegister()
	uniqueId := gameConfig.Identifier
	initState := gameConfig.InitState

	// Make a gameState
	gameRenderState := shared.GameRenderState{
		PlayerLoc:shared.Coord{3,3},
		OtherPlayers: make(map[string]shared.Coord),
		Prey: shared.Coord{5,5},
	}

	// Create player node
	pn := PlayerNode{
		pixelInterface: pixelInterface,
		nodeInterface: nodeInterface,
		playerCommChannel: playerCommChannel,
		geo: geometry.CreateNewGridManager(initState.Settings),
		GameRenderState: gameRenderState,
		identifier: uniqueId,
		GameConfig: initState,
	}

	// Allow the node-node interface to refer back to this node
	nodeInterface.PlayerNode = &pn

	return pn
}


func (pn * PlayerNode) RunGame() {

	for {
		message := <-pn.playerCommChannel
		switch message {
		case "quit":
			break
		default:
			pn.movePlayer(message)
			pn.pixelInterface.SendPlayerGameState(pn.GameRenderState)
			fmt.Println("movin' player", message)
		}
	}

}

func (pn * PlayerNode) movePlayer(move string) {
	// Get current player state
	playerLoc := pn.GameRenderState.PlayerLoc

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
	if pn.geo.IsValidMove(newPosition) {
		pn.GameRenderState.PlayerLoc = newPosition
	}
}
