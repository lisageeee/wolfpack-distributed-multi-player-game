package impl

import (
	"net"
	"../../shared"
	"encoding/json"
	"fmt"
)

// The interface with the player's Pixel GUI (pixel-node.go) from the logic node
type PixelInterface struct {
	pixelListener     *net.UDPConn
	pixelWriter 	  *net.UDPConn
	playerCommChannel chan string
	playerSendChannel chan shared.GameState
	Id string
}

// Creates & returns a pixel interface with a channel to send string information to the main node over
// Called by the main logic node package
func CreatePixelInterface(playerCommChannel chan string, playerSendChannel chan shared.GameState, id string) PixelInterface {
	pi := PixelInterface{playerCommChannel: playerCommChannel,playerSendChannel:playerSendChannel, Id: id}
	return pi
}

func (pi *PixelInterface) waitForGameStates() {
	for {
		state := <-pi.playerSendChannel
		// Create the player map without without this node or prey node
		otherPlayers := make(map[string]shared.Coord)
		for key, value := range state.PlayerLocs {
			if key != pi.Id && key != "prey" {
				otherPlayers[key] = value
			}
		}

		renderState := shared.GameRenderState{
			PlayerLoc:    state.PlayerLocs[pi.Id],
			Prey:         state.PlayerLocs["prey"],
			OtherPlayers: otherPlayers,
		}

		toSend, err := json.Marshal(renderState)
		if err != nil {
			fmt.Println(err)
		} else {
			// Send position to player node
			fmt.Println("sending position")
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
func (pi * PixelInterface) RunPlayerListener(sendingAddr string, receivingAddr string) {

	_, playerInput := StartListenerUDP(receivingAddr)
	defer playerInput.Close()

	playerSend := StartSenderUDP(sendingAddr)
	defer playerSend.Close()

	pi.pixelWriter = playerSend
	pi.pixelListener = playerInput

	go pi.waitForGameStates()

	// takes a listener client
	// runs the listener in a infinite loop
	player := pi.pixelListener
	player.SetReadBuffer(1024)

	for {
		buf := make([]byte, 1024)
		rlen, _, err := player.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		} else {
			// Write to comm channel for node to receive
			pi.playerCommChannel <- string(buf[0:rlen])
		}
	}
}
