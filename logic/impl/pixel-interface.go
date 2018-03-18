package impl

import (
	"net"
	"../../shared"
	"encoding/json"
	"fmt"
)

type PixelInterface struct {
	pixelListener     *net.UDPConn
	pixelWriter 	  *net.UDPConn
	playerCommChannel chan string
}

func CreatePixelInterface(playerCommChannel chan string) PixelInterface {
	pi := PixelInterface{playerCommChannel: playerCommChannel}
	return pi
}

func (pi *PixelInterface) SendPlayerGameState(state shared.GameRenderState) {
	// TODO: add other nodes
	toSend, err := json.Marshal(state)
	if err != nil {
		fmt.Println(err)
	} else {
		// Send position to player node
		fmt.Println("sending position")
		pi.pixelWriter.Write([]byte(toSend))
	}
}

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
	player.SetReadBuffer(1048576)

	for {
		buf := make([]byte, 1024)
		rlen, _, err := player.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		} else {
			pi.playerCommChannel <- string(buf[0:rlen])
		}
		// Write to comm channel for node to deal with

	}
}
