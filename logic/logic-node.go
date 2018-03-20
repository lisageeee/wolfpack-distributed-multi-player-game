package main

import (
	"fmt"
	"os"
	_ "image/png"
	_ "image/jpeg"
	l "./impl"
	"../key-helpers"
)

// Entrypoint, sets up communication channels and creates the RemotePlayerInterface
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
	node := l.CreatePlayerNode(node_listener_ip_address, player_listener_ip_address, pixel_ip_address, *pubKey, *privKey)
	node.RunGame()
}
