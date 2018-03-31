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
	pixelListener     *net.TCPListener
	pixelWriter 	  *net.TCPConn
	playerCommChannel chan string
	playerSendChannel chan shared.GameState
	gameConfig		  shared.InitialGameSettings
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

func (pi *PixelInterface) waitForGameStates() {
	for {
		state := <-pi.playerSendChannel
		// Create the player map without without this node or prey node
		otherPlayers := make(map[string]shared.Coord)
		state.PlayerLocs.RLock()
		for key, value := range state.PlayerLocs.Data {
			if key != pi.Id && key != "prey" {
				otherPlayers[key] = value
			}
		}

		renderState := shared.GameRenderState{
			PlayerLoc:    state.PlayerLocs.Data[pi.Id],
			Prey:         state.PlayerLocs.Data["prey"],
			OtherPlayers: otherPlayers,
		}

		state.PlayerLocs.RUnlock()

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
// pixel-node, respectively. Must be run in a goroutine (infinite loop_
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
		} else {
			// Write to comm channel for node to receive
			pi.playerCommChannel <- string(buf[0:rlen])
		}
	}
}

func (pi * PixelInterface) GetTCPConn() (*net.TCPConn) {
	// gets the initial TCP conn
	player := pi.pixelListener
	conn, err := player.AcceptTCP()
	if err != nil {
		fmt.Println(err)
	}
	// Send the pixel node gameConfig immediately
	marshalledConfig, err := json.Marshal(&pi.gameConfig)
	if err != nil {
		fmt.Println(err)
	} else {
		conn.Write(marshalledConfig)
	}
	return conn
}
