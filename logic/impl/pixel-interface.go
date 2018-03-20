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
}

// Creates & returns a pixel interface with a channel to send string information to the main node over
// Called by the main logic node package
func CreatePixelInterface(playerCommChannel chan string) PixelInterface {
	pi := PixelInterface{playerCommChannel: playerCommChannel}
	return pi
}

// Sends a game state to the player's pixel interface for rendering
func (pi *PixelInterface) SendPlayerGameState(state shared.GameRenderState) {
	toSend, err := json.Marshal(state)
	if err != nil {
		fmt.Println(err)
	} else {
		// Send position to player node
		fmt.Println("sending position")
		pi.pixelWriter.Write([]byte(toSend))
	}
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
