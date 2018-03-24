package impl

import (
	"../../shared"
	"../../geometry"
	"fmt"
	"crypto/ecdsa"
)

// The "main" node part of the logic node. Deals with computation and checks; not communications
type PlayerNode struct {
	pixelInterface	  PixelInterface
	nodeInterface 	  *NodeCommInterface
	playerCommChannel chan string
	GameState		  shared.GameState
	//GameRenderState	  shared.GameRenderState
	geo        geometry.GridManager
	Identifier string
	GameConfig shared.InitialState
}

// Creates the main logic node and required interfaces with the arguments passed in logic-node.go
// nodeListenerAddr = where we expect to receive messages from other nodes
// playerListenerAddr = where we expect to receive messages from the pixel-node
// pixelSendAddr = where we will be sending new game states to the pixel node
func CreatePlayerNode(nodeListenerAddr, playerListenerAddr, pixelSendAddr string, pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey) (PlayerNode) {
	// Setup the player communication buffered channel
	playerCommChannel := make(chan string, 5)

	// Startup Pixel interface + listening
	pixelInterface := CreatePixelInterface(playerCommChannel)
	go pixelInterface.RunPlayerListener(pixelSendAddr, playerListenerAddr)

	// Start the node to node interface
	nodeInterface := CreateNodeCommInterface(pubKey, privKey)
	addr, listener := StartListenerUDP(nodeListenerAddr)
	nodeInterface.LocalAddr = addr
	nodeInterface.IncomingMessages = listener
	go nodeInterface.RunListener(listener, nodeListenerAddr)

	// Register with server, update info
	uniqueId := nodeInterface.ServerRegister()
	go nodeInterface.SendHeartbeat()

	// TODO: This starts a race condition. We might need to either
	// block or get a unique id earlier



	//// Make a gameState
	playerLocs := make(map[string]shared.Coord)
	playerLocs["prey"] = shared.Coord{5,5}
	playerLocs[uniqueId] = shared.Coord{3,3}

	// Make a gameState
	gameState := shared.GameState{
		PlayerLocs: playerLocs,
	}

	// Create player node
	pn := PlayerNode{
		pixelInterface:    pixelInterface,
		nodeInterface:     &nodeInterface,
		playerCommChannel: playerCommChannel,
		geo:               geometry.CreateNewGridManager(nodeInterface.Config.InitState.Settings),
		GameState:         gameState,
		Identifier:        uniqueId,
		GameConfig:        nodeInterface.Config.InitState,
	}

	// Allow the node-node interface to refer back to this node
	nodeInterface.PlayerNode = &pn

	return pn
}

// Runs the main node (listens for incoming messages from pixel interface) in a loop, must be called at the
// end of main (or alternatively, in a goroutine)
func (pn * PlayerNode) RunGame() {

	for {
		message := <-pn.playerCommChannel
		switch message {
		case "quit":
			break
		default:
			move := pn.movePlayer(message)
			pn.pixelInterface.SendPlayerGameState(pn.GameState, pn.Identifier)
			pn.nodeInterface.SendMoveToNodes(&move)
			fmt.Println("movin' player", message)
		}
	}

}

// Given a string "up"/"down"/"left"/"right", changes the player state to make that move iff that move is valid
// (not into a wall, out of bounds)
func (pn * PlayerNode) movePlayer(move string) (shared.Coord) {
	// Get current player state
	playerLoc := pn.GameState.PlayerLocs[pn.Identifier]

	originalPosition := shared.Coord{X: playerLoc.X, Y: playerLoc.Y}
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
	if pn.geo.IsValidMove(newPosition) && pn.geo.IsNotTeleporting(originalPosition, newPosition){
		pn.GameState.PlayerLocs[pn.Identifier] = newPosition
		return newPosition
	}
	return playerLoc
}

// GETTERS

func (pn *PlayerNode) GetPixelInterface() (PixelInterface) {
	return pn.pixelInterface
}

func (pn *PlayerNode) GetNodeInterface() (*NodeCommInterface) {
	return pn.nodeInterface
}
