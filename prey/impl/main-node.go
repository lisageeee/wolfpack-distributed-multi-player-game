package impl

import (
	"../../shared"
	"../../geometry"
	"crypto/ecdsa"
	"time"
	"math/rand"
	"fmt"
)

// The "main" node part of the logic node. Deals with computation and checks; not communications
type PreyNode struct {

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
	go nodeInterface.ManageOtherNodes()
	go nodeInterface.PruneNodes()
	// Register with server, update info
	uniqueId := nodeInterface.ServerRegister()
	go nodeInterface.SendHeartbeat()

	// Make a gameState
	playerLocs := make(map[string]shared.Coord)
	playerLocs[uniqueId] = shared.Coord{5,5}
	playerMap := shared.PlayerLockMap{Data:playerLocs}

	playerScores := make(map[string]int)
	playerScoreMap := shared.ScoresLockMap{Data:playerScores}

	// Make a gameState
	gameState := shared.GameState{
		PlayerLocs: playerMap,
		PlayerScores: playerScoreMap,
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
	ticker := time.NewTicker(time.Millisecond * 250)
	for _ = range ticker.C {
		var dir string
		randMove := rand.Float64()
		if randMove < 0.25 || len(pn.GameState.PlayerLocs.Data) <2{
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

				move := pn.MovePrey(dir)
				pn.nodeInterface.SendMoveToNodes(&move)
			}()
		}else {
			go func() {
				dir := "still"
				maxVal := -1
				for _,i:= range []int{-1,1}{
					val := pn.MaxDistance(pn.GameState.PlayerLocs, i, 0)
						if val > maxVal {
							maxVal = val
							if i == -1{
								dir = "left"
							}else{
								dir = "right"
							}

					}

				}
				for _,j:= range []int{-1,1}{
					val := pn.MaxDistance(pn.GameState.PlayerLocs, 0, j)
					if val > maxVal{
						maxVal = val
						if j == -1{
							dir = "down"
						}else{
							dir = "up"
						}
					}
				}

				fmt.Println("move prey: ", dir)
				move := pn.MovePrey(dir)
				pn.nodeInterface.SendMoveToNodes(&move)
			}()
		}
	}
}

func (pn *PreyNode)MaxDistance(playerLocs shared.PlayerLockMap, moveX, moveY int)int {
	playerLocs.RLock()
	myLoc := playerLocs.Data["prey"]
	newLoc := shared.Coord{myLoc.X+moveX, myLoc.Y+moveY}
	if !pn.geo.IsValidMove(newLoc)	{
		return -1
	}
	dist := 0
	for val, nodeLoc := range playerLocs.Data {
		if val != "prey"{
			dist += abs(newLoc.X-nodeLoc.X)+ abs(newLoc.Y-nodeLoc.Y)
		}
	}
	playerLocs.RUnlock()
	return dist
}
func abs(num int)int {
	if num <0{
		return -num
	}else{
		return num
	}
}

func (pn * PreyNode) MovePrey(move string) (shared.Coord) {
	pn.GameState.PlayerLocs.RLock()
	preyLoc := pn.GameState.PlayerLocs.Data["prey"]
	pn.GameState.PlayerLocs.RUnlock()

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
		pn.GameState.PlayerLocs.Lock()
		pn.GameState.PlayerLocs.Data["prey"] = newPosition
		pn.GameState.PlayerLocs.Unlock()
		return newPosition
	}
	return preyLoc
}

// GETTERS

func (pn *PreyNode) GetNodeInterface() (*NodeCommInterface) {
	return pn.nodeInterface
}
func (pn *PreyNode) GetGridManager() (*geometry.GridManager) {
	return &pn.geo
}