package impl

import (
	"../../shared"
	"../../geometry"
	"fmt"
	"crypto/ecdsa"
	"time"
	"math/rand"
)

// The "main" node part of the logic node. Deals with computation and checks; not communications
type PreyNode struct {
	nodeInterface 	  *NodeCommInterface
	playerCommChannel chan string
	//playerSendChannel chan shared.GameState
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
func CreatePreyNode(nodeListenerAddr, playerListenerAddr string,
	pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey, serverAddr string) (PreyNode) {
	// Setup the player communication buffered channel
	playerCommChannel := make(chan string, 5)
	//playerSendChannel := make(chan shared.GameState, 5)

	// Start the node to node interface
	nodeInterface := CreateNodeCommInterface(pubKey, privKey, serverAddr)
	addr, listener := StartListenerUDP(nodeListenerAddr)
	nodeInterface.LocalAddr = addr
	nodeInterface.IncomingMessages = listener
	go nodeInterface.RunListener(listener, nodeListenerAddr)

	//// Register with server, update info
	uniqueId := nodeInterface.ServerRegister()
	go nodeInterface.SendHeartbeat()

	//// Make a gameState
	playerLocs := make(map[string]shared.Coord)
	playerLocs["prey"] = shared.Coord{5,5}
	//playerLocs[uniqueId] = shared.Coord{3,3}

	// Make a gameState
	gameState := shared.GameState{
		PlayerLocs: playerLocs,
	}

	// Create player node
	pn := PreyNode{
		nodeInterface:     &nodeInterface,
		playerCommChannel: playerCommChannel,
		//playerSendChannel:playerSendChannel,
		geo:               geometry.CreateNewGridManager(nodeInterface.Config.InitState.Settings),
		GameState:         gameState,
		Identifier:        uniqueId,
		GameConfig:        nodeInterface.Config.InitState,
	}

	// Allow the node-node interface to refer back to this node
	nodeInterface.PreyNode = &pn

	return pn
}

// Runs the main node (listens for incoming messages from pixel interface) in a loop, must be called at the
// end of main (or alternatively, in a goroutine)
func (pn * PreyNode) RunGame(playerListener string) {
	//for {
	//	message := <-pn.playerCommChannel
	//	switch message {
	//	case "quit":
	//		break
	//	default:
	//		move := pn.movePrey(message)
	//		pn.nodeInterface.SendMoveToNodes(&move)
	//		fmt.Println("movin' player", message)
	//	}
	//}
	fmt.Println("Hello")
	ticker := time.NewTicker(time.Millisecond * 100)
	for _ = range ticker.C {
		var dir string

		go func() {
				random := rand.Float64()
				switch {
				case random < 0.25:
					dir = "up"
				case random < 0.5:
					dir = "down"
				case random < 0.75:
					dir = "right"
				default:
					dir = "left"
				}

				move := pn.movePrey(dir)
				pn.nodeInterface.SendMoveToNodes(&move)
				//fmt.Println(newVec)
				//node.GameState.Prey = shared.Coord{int(preyPos.X), int(preyPos.Y)}
				//node.RenderNewState(win)
				//node.GameState.Prey = newVec
				//node.RenderNewState(win)
				//win.Update()
		}()
	}
}

// Given a string "up"/"down"/"left"/"right", changes the player state to make that move iff that move is valid
// (not into a wall, out of bounds)
func (pn * PreyNode) movePrey(move string) (shared.Coord) {
	// Get current player state
	playerLoc := pn.GameState.PlayerLocs["prey"]

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

func (pn *PreyNode) GetNodeInterface() (*NodeCommInterface) {
	return pn.nodeInterface
}
