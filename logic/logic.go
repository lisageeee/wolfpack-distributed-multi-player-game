package main

import (
	"fmt"
	"os"
	_ "image/png"
	_ "image/jpeg"
	logicImpl "./impl"
	"../key-helpers"
)

// Entrypoint for the player (logic) node, creates the node and all interfaces by calling the playerNode constructor
// and calling runGame
func main() {
	fmt.Println("hello world")

	// Default IP addresses if none provided
	nodeListenerAddr := ":0"
	playerListenerIpAddress := ":12345"
	serverAddr := ":8081"
	// Can start with an IP as param
	if len(os.Args) > 3 {
		nodeListenerAddr = os.Args[1]
		playerListenerIpAddress = os.Args[2]
		serverAddr = os.Args[3]
	} else if len(os.Args) > 2 {
		nodeListenerAddr = os.Args[1]
		playerListenerIpAddress = os.Args[2]
	} else if len(os.Args)>1{
		nodeListenerAddr = ":0"
		playerListenerIpAddress = os.Args[1]
	}

	pubKey, privKey := key_helpers.GenerateKeys()
	node := logicImpl.CreatePlayerNode(nodeListenerAddr, playerListenerIpAddress, pubKey, privKey, serverAddr)
	node.RunGame(playerListenerIpAddress)
}
