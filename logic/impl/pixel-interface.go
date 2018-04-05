package impl

import (
	"net"
	"../../shared"
	"encoding/json"
	"fmt"
	"log"
)

// The interface with the player's Pixel GUI (pixel-node.go) from the logic node
type PixelInterface struct {
	// The TCP connection over which we wait for a pixel node to connect
	pixelListener     *net.TCPListener

	// The connection over which messages are sent after a connection is established
	pixelWriter 	  *net.TCPConn

	// The channel used to send player moves to the main node
	playerCommChannel chan string

	// The channel used to write new game states that should be sent to the pixel node to
	playerSendChannel chan shared.GameState

	// The gameconfig for this game
	gameConfig		  shared.InitialGameSettings

	// The ID of this logic node
	Id string
}

// Creates & returns a pixel interface with a channel to send string information to the main node over
// Called by the main logic node package
func CreatePixelInterface(playerCommChannel chan string, playerSendChannel chan shared.GameState,
	settings shared.InitialGameSettings, id string) PixelInterface {
	pi := PixelInterface{playerCommChannel: playerCommChannel,playerSendChannel:playerSendChannel, Id: id,
	gameConfig: settings}
	return pi
}

// To be run in a goroutine; waits for the notification a gamestate should be rendered then sends that gamestate
// to the pixel node
func (pi *PixelInterface) waitForGameStates() {
	for {
		state := <-pi.playerSendChannel


		state.PlayerLocs.Lock()
		state.PlayerScores.Lock()

		// Create the player map without without this node or prey node
		otherPlayers := make(map[string]shared.Coord)
		for key, value := range state.PlayerLocs.Data {
			if key != pi.Id && key != "prey" {
				otherPlayers[key] = value
			}
		}

		otherScores := make(map[string]int)
		for key, value := range state.PlayerScores.Data {
			if key != pi.Id {
				otherScores[key] = value
			} else {
				otherScores["ME"] = value
			}
		}

		renderState := shared.GameRenderState{
			PlayerLoc:    state.PlayerLocs.Data[pi.Id],
			Prey:         state.PlayerLocs.Data["prey"],
			OtherPlayers: otherPlayers,
			Scores: otherScores,
		}

		state.PlayerScores.Unlock()
		state.PlayerLocs.Unlock()

		toSend, err := json.Marshal(renderState)
		if err != nil {
			fmt.Println(err)
		} else {
			// Send position to player node
			pi.pixelWriter.Write(toSend)
		}
	}
}
// Sends a game state to the player's pixel interface for rendering
func (pi *PixelInterface) SendPlayerGameState(state shared.GameState) {
	pi.playerSendChannel <- state
}

// Given two local UDP addresses, initializes the ports for sending and receiving messages from the
// pixel-node, respectively. Must be run in a goroutine (infinite loop)
func (pi * PixelInterface) RunPlayerListener(receivingAddr string) {

	addr, _ := net.ResolveTCPAddr("tcp",receivingAddr)
	playerInput, _ := net.ListenTCP("tcp", addr)
	pi.pixelListener = playerInput
	fmt.Println("about to get to conn")
	conn := pi.GetTCPConn()
	fmt.Println("got conn")
	pi.pixelWriter = conn

	go pi.waitForGameStates()

	// takes a listener client
	// runs the listener in a infinite loop
	player := pi.pixelWriter
	for {
		buf := make([]byte, 1024)
		rlen, err := player.Read(buf)
		if err != nil {
			log.Fatal("Pixel node disconnected")
		} else if string(buf[0:rlen]) == "getgameconfig"{
			SendGameConfig(pi, player)
		} else {
			// Write to comm channel for node to receive
			pi.playerCommChannel <- string(buf[0:rlen])
		}
	}
}

// Listener function that waits for a pixel node to connect, returns the resulting TCPConn after connection.
func (pi * PixelInterface) GetTCPConn() (*net.TCPConn) {
	// gets the initial TCP conn
	player := pi.pixelListener
	conn, err := player.AcceptTCP()
	if err != nil {
		fmt.Println(err)
	}
	SendGameConfig(pi, conn)
	return conn
}

func SendGameConfig(pi *PixelInterface, conn *net.TCPConn) {
	// Send the pixel node gameConfig immediately
	marshalledConfig, err := json.Marshal(&pi.gameConfig)
	if err != nil {
		fmt.Println(err)
	} else {
		conn.Write(marshalledConfig)
	}
}
