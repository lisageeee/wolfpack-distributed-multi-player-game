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

	// Listener IP address
	var node_listener_ip_address string
	var player_listener_ip_address string
	var pixel_ip_address string
	// Can start with an IP as param
	if len(os.Args) > 2 {
		node_listener_ip_address = os.Args[1]
		player_listener_ip_address = os.Args[2]
	} else if len(os.Args)>1{
		node_listener_ip_address = "127.0.0.1:0"
		player_listener_ip_address = os.Args[1]
		pixel_ip_address = "127.0.0.1:1234"
	} else {
		node_listener_ip_address = "127.0.0.1:0"
		player_listener_ip_address = "127.0.0.1:12345"
		pixel_ip_address = "127.0.0.1:1234"
	}

	pubKey, privKey := key_helpers.GenerateKeys()
	node := logicImpl.CreatePlayerNode(node_listener_ip_address, player_listener_ip_address, pixel_ip_address, *pubKey, *privKey)
	node.RunGame()
}
