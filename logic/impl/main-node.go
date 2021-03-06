package impl

import (
	"../../shared"
	"../../geometry"
	"fmt"
	"crypto/ecdsa"
	"time"
	"math"
)

// The "main" node part of the logic node. Deals with computation and checks; not communications
type PlayerNode struct {

	// The interface that deals with incoming and outgoing messages from the associated Pixel node
	pixelInterface	  PixelInterface

	// The interface that deals with incoming and outgoing messages from other logic nodes
	nodeInterface 	  *NodeCommInterface

	// Channel on which incoming player moves will be passed from the pixelInterface to the logic node
	playerCommChannel chan string

	// Channel on which outgoing player states will be passed from this playerNode to the pixelInterface for sending
	// to the player
	playerSendChannel chan shared.GameState

	// The current gamestate, represented as a map of player identifiers to locations
	GameState		  shared.GameState

	// The grid manager for the current game, which determines valid moves
	geo        geometry.GridManager

	// This logic node's identifier, assigned upon registration with the server
	Identifier string

	// The game configuration provided upon registration from the server. Includes wall locations and board size.
	GameConfig shared.InitialState
}

// Creates the main logic node and required interfaces with the arguments passed in logic-node.go
// nodeListenerAddr = where we expect to receive messages from other nodes
// playerListenerAddr = where we expect to receive messages from the pixel-node
// pixelSendAddr = where we will be sending new game states to the pixel node
func CreatePlayerNode(nodeListenerAddr, playerListenerAddr string,
	pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey, serverAddr string) (PlayerNode) {
	// Setup the player communication buffered channel
	playerCommChannel := make(chan string, 5)
	playerSendChannel := make(chan shared.GameState, 5)

	// Start the node to node interface
	nodeInterface := CreateNodeCommInterface(pubKey, privKey, serverAddr)
	addr, listener := StartListenerUDP(nodeListenerAddr)

	nodeInterface.LocalAddr = addr
	nodeInterface.IncomingMessages = listener
	go nodeInterface.RunListener(listener, nodeListenerAddr)
	go nodeInterface.ManageOtherNodes()
	go nodeInterface.ManageAcks()
	go nodeInterface.PruneNodes()
	go nodeInterface.SendGameStateToPixel()

	// Register with server, update info
	uniqueId := nodeInterface.ServerRegister()
	go nodeInterface.SendHeartbeat()

	// Startup Pixel interface + listening
	pixelInterface := CreatePixelInterface(playerCommChannel, playerSendChannel,
		nodeInterface.Config.InitState.Settings, uniqueId)

	//// Make a gameState
	playerLocs := make(map[string]shared.Coord)
	playerLocs["prey"] = shared.Coord{5,5}
	playerLocs[uniqueId] = shared.Coord{1,1}

	playerScores := make(map[string]int)
	playerScores[uniqueId] = 0

	playerMap := shared.PlayerLockMap{Data:playerLocs}
	scoreMap := shared.ScoresLockMap{Data:playerScores}

	// Make a gameState
	gameState := shared.GameState{
		PlayerLocs: playerMap,
		PlayerScores: scoreMap,
	}

	// Create player node
	pn := PlayerNode{
		pixelInterface:    pixelInterface,
		nodeInterface:     &nodeInterface,
		playerCommChannel: playerCommChannel,
		playerSendChannel: playerSendChannel,
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
func (pn * PlayerNode) RunGame(playerListener string) {
	go pn.pixelInterface.RunPlayerListener(playerListener)
	fmt.Println("listener running")

	for {
		message := <-pn.playerCommChannel
		switch message {
		case "quit":
			break
		default:
			move, didMove := pn.movePlayer(message)
			if didMove {
				pn.nodeInterface.SendMoveToNodes(&move)
			}
			if pn.nodeInterface.CheckGotPrey(move) == nil {
				fmt.Println("Got the prey")
				pn.GameState.PlayerScores.Lock()
				pn.GameState.PlayerScores.Data[pn.Identifier] += pn.GameConfig.CatchWorth
				pn.nodeInterface.SendPreyCaptureToNodes(&move, pn.GameState.PlayerScores.Data[pn.Identifier])
				pn.nodeInterface.RW.Add("captured_prey", sequenceNumber, &move)
				fmt.Println(pn.GameState.PlayerScores.Data[pn.Identifier])
				pn.GameState.PlayerScores.Unlock()
			}
			// pn.pixelInterface.SendPlayerGameState(pn.GameState)
		}
	}

}

// Given a string "up"/"down"/"left"/"right", changes the player state to make that move iff that move is valid
// (not into a wall, out of bounds)
func (pn * PlayerNode) movePlayer(move string) (newPos shared.Coord, changed bool) {
	// Get current player state
	pn.GameState.PlayerLocs.RLock()
	playerLoc := pn.GameState.PlayerLocs.Data[pn.Identifier]
	pn.GameState.PlayerLocs.RUnlock()

	originalPosition := shared.Coord{X: playerLoc.X, Y: playerLoc.Y}
	// Calculate new position with move
	newPosition := shared.Coord{X: playerLoc.X, Y: playerLoc.Y}
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
		//pn.GameState.PlayerLocs.Lock()
		//pn.GameState.PlayerLocs.Data[pn.Identifier] = newPosition
		//pn.GameState.PlayerLocs.Unlock()
		return newPosition, true
	}
	return playerLoc, false
}

// GETTERS
// Returns the pixel interface; mainly of use for testing
func (pn *PlayerNode) GetPixelInterface() (PixelInterface) {
	return pn.pixelInterface
}
// Returns nothe node interface; mainly of use for testing
func (pn *PlayerNode) GetNodeInterface() (*NodeCommInterface) {
	return pn.nodeInterface
}

// Return the grid manager of this player node
func (pn *PlayerNode) GetGridManager() (*geometry.GridManager) {
	return &pn.geo
}

// Return the comm channel used to communicate with the pixel interface, mostly for testing
func (pn *PlayerNode) GetPlayerCommChannel() (chan string) {
	return pn.playerCommChannel
}

// Runs a bot game
func (pn * PlayerNode) RunBotGame(playerListener string) {
	for {
		myState := pn.GameState.PlayerLocs.Data[pn.Identifier]
		prey := pn.GameState.PlayerLocs.Data["prey"]
		command := "still"
		minVal := abs(myState.X-prey.X)+ abs(myState.Y-prey.Y)
		if minVal <= 3{
			minVal = math.MaxInt8
		}
		for _,i:= range []int{-1,1}{
			val := abs(myState.X+i-prey.X)+ abs(myState.Y-prey.Y)
			if val < minVal && pn.geo.IsValidMove(shared.Coord{myState.X+i, myState.Y}) {
				minVal = val
				if i == -1{
					command = "left"
				}else{
					command = "right"
				}
			}
		}
		for _,j:= range []int{-1,1}{
			val := abs(myState.X-prey.X)+ abs(myState.Y+j-prey.Y)
			if val < minVal &&  pn.geo.IsValidMove(shared.Coord{myState.X, myState.Y+j}) {
				minVal = val
				if j == -1{
					command = "down"
				}else{
					command = "up"
				}
			}
		}
		move, ok := pn.movePlayer(command)
		if ok{
			pn.nodeInterface.SendMoveToNodes(&move)
			pn.nodeInterface.GameStateToSend = make(chan bool, 30)
			fmt.Println("movin' bot", command)
		}
		if pn.nodeInterface.CheckGotPrey(move) == nil {
			fmt.Println("Got the prey")
			pn.GameState.PlayerScores.Lock()
			pn.GameState.PlayerScores.Data[pn.Identifier] += pn.GameConfig.CatchWorth
			pn.nodeInterface.SendPreyCaptureToNodes(&move, pn.GameState.PlayerScores.Data[pn.Identifier])
			pn.nodeInterface.RW.Add("captured_prey", sequenceNumber, &move)
			fmt.Println(pn.GameState.PlayerScores.Data[pn.Identifier])
			pn.GameState.PlayerScores.Unlock()
		}
		// Take move off the channel
		time.Sleep(time.Millisecond*400)
	}
}
func abs(num int)int {
	if num <0{
		return -num
	}else{
		return num
	}
}