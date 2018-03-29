package impl

import (
	"../../shared"
	"../../geometry"
	"crypto/ecdsa"
	"time"
	"math/rand"
)

// The "main" node part of the logic node. Deals with computation and checks; not communications
type PreyNode struct {
	nodeInterface 	  *NodeCommInterface
	playerCommChannel chan string
	GameState		  shared.GameState
	geo        geometry.GridManager
	Identifier string
	GameConfig shared.InitialState
}

// nodeListenerAddr = where we expect to receive messages from other nodes
func CreatePreyNode(nodeListenerAddr, playerListenerAddr string,
	pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey, serverAddr string) (PreyNode) {
	// Setup the player communication buffered channel
	playerCommChannel := make(chan string, 5)

	// Start the node to node interface
	nodeInterface := CreateNodeCommInterface(pubKey, privKey, serverAddr)
	addr, listener := StartListenerUDP(nodeListenerAddr)
	nodeInterface.LocalAddr = addr
	nodeInterface.IncomingMessages = listener
	go nodeInterface.RunListener(listener, nodeListenerAddr)

	// Register with server, update info
	uniqueId := nodeInterface.ServerRegister()
	go nodeInterface.SendHeartbeat()

	// Make a gameState
	playerLocs := make(map[string]shared.Coord)
	playerLocs[uniqueId] = shared.Coord{5,5}

	// Make a gameState
	gameState := shared.GameState{
		PlayerLocs: playerLocs,
	}

	// Create Prey node
	pn := PreyNode{
		nodeInterface:     &nodeInterface,
		playerCommChannel: playerCommChannel,
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
		}()
	}
}


func (pn * PreyNode) movePrey(move string) (shared.Coord) {
	preyLoc := pn.GameState.PlayerLocs["prey"]

	originalPosition := shared.Coord{X: preyLoc.X, Y: preyLoc.Y}

	newPosition := shared.Coord{X: preyLoc.X, Y: preyLoc.Y}

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
	// Check new move is valid, if so update prey position
	if pn.geo.IsValidMove(newPosition) && pn.geo.IsNotTeleporting(originalPosition, newPosition){
		pn.GameState.PlayerLocs["prey"] = newPosition
		return newPosition
	}
	return preyLoc
}

// GETTERS

func (pn *PreyNode) GetNodeInterface() (*NodeCommInterface) {
	return pn.nodeInterface
}
