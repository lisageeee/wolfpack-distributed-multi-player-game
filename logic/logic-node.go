package main

import (
	"fmt"
	"net"
	"os"
	"net/rpc"
	"strconv"
	_ "image/png"
	_ "image/jpeg"
	"log"
	"../shared"
	"../geometry"
	l "./impl"
	"encoding/json"
)

type RemotePlayerInterface struct {
	pixelInterface	  l.PixelInterface
	playerCommChannel chan string
	otherNodes        *net.UDPConn
	GameRenderState	  shared.GameRenderState
	geo               geometry.GridManager
	identifier        int
	GameConfig		  shared.InitialState
}

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

	// Setup the player communcation channel
	playerCommChannel := make(chan string)

	// Startup Pixel interface + listening
	pixelInterface := l.CreatePixelInterface(playerCommChannel)
	go pixelInterface.RunPlayerListener(pixel_ip_address, player_listener_ip_address)

	// Start the node to node interface
	_, client := StartListener(node_listener_ip_address)
	defer client.Close()

	gameConfig := ServerRegister(client.LocalAddr().String())
	otherNodes := gameConfig.Connections
	uniqueId := gameConfig.Identifier
	fmt.Println("Your identifier is:")
	fmt.Println(uniqueId)
	fmt.Println("The connections:")
	fmt.Println(otherNodes)
	initState := gameConfig.InitState
	udpAddr := client.LocalAddr().(*net.UDPAddr)


	// Make default gameState
	gameRenderState := shared.GameRenderState{
		PlayerLoc: shared.Coord{3, 3},
		OtherPlayers: make(map[string]shared.Coord),
		Prey: shared.Coord{5, 5}} // TODO: change these to dynamic when we connect to other players/prey

	pi := RemotePlayerInterface{
		pixelListener: player,
		pixelWriter: pixel,
		otherNodes: client,
		playerCommChannel: make(chan string),
		geo: geometry.CreateNewGridManager(initState.Settings),
		GameRenderState: gameRenderState,
		identifier: uniqueId,
		GameConfig: initState}
	floodNodes(otherNodes, udpAddr, pi)
	pi.runGame()
}

func (pi *RemotePlayerInterface) runGame() {
	go pi.runNodeListener()

	for {
		message := <-pi.playerCommChannel
		switch message {
		case "quit":
			break
		default:
			pi.movePlayer(message)
			pi.pixelInterface.SendPlayerGameState(pi.GameRenderState)
			fmt.Println("movin' player", message)
		}
	}

}

func (pi * RemotePlayerInterface) movePlayer(move string) {
	// Get current player state
	playerLoc := pi.GameRenderState.PlayerLoc

	// Calculate new position with move
	newPosition := shared.Coord{X: playerLoc.X, Y: playerLoc.Y}
	fmt.Println(move)
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
	if pi.geo.IsValidMove(newPosition) {
		pi.GameRenderState.PlayerLoc = newPosition
	}
}


func StartListener(ip_addr string) (*net.UDPAddr, *net.UDPConn) {
	// takes an ip address and port to listen on
	// returns the udp address and listener client
	// starts Listener
	udp_addr, _ := net.ResolveUDPAddr("udp", ip_addr)
	client, err := net.ListenUDP("udp", udp_addr)
	if err != nil {
		panic(err)
	}
	return udp_addr, client
}

const udp_generic = "127.0.0.1:0"
var clients []*net.Conn
func (pi * RemotePlayerInterface) runNodeListener() {
	// takes a listener client
	// runs the listener in a infinite loop
	client := pi.otherNodes

	client.SetReadBuffer(1048576)

	i := 0
	for {
		i++
		buf := make([]byte, 1024)
		rlen, addr, err := client.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(buf[0:rlen]))
		fmt.Println(addr)
		fmt.Println(i)
		if "sgs" == string(buf[0:3]){
			id := string(buf[3])
			if err != nil{
				panic(err)
			}
			var remoteCoord shared.Coord
			err2 := json.Unmarshal(buf[4:rlen], &remoteCoord)
			if err2 != nil {
				panic(err)
			}
			pi.GameRenderState.OtherPlayers[id] = remoteCoord
			fmt.Println(pi.GameRenderState.OtherPlayers)
		}else if string(buf[0:rlen]) != "connected" {
			remote_client, err := net.Dial("udp", string(buf[0:rlen]))
			if err != nil {
				panic(err)
			}
			toSend, err := json.Marshal(pi.GameRenderState.PlayerLoc)
			// Code sgs sends the connecting node the gamestate
			remote_client.Write([]byte("sgs" + strconv.Itoa(pi.identifier) + string(toSend)))
			clients = append(clients, &remote_client)
		}

	}
}

func floodNodes(otherNodes []string, udp_addr *net.UDPAddr, pi RemotePlayerInterface) {
	localIP, _ := net.ResolveUDPAddr("udp", udp_generic)
	for _, ip := range otherNodes {
		node_udp, _ := net.ResolveUDPAddr("udp", ip)
		// Connect to other node
		node_client, err := net.DialUDP("udp", localIP, node_udp)
		if err != nil {
			panic(err)
		}
		// Exchange messages with other node
		myListener := udp_addr.IP.String() + ":" + strconv.Itoa(udp_addr.Port)
		node_client.Write([]byte(myListener))
		initial_game, err := json.Marshal(pi.GameRenderState.PlayerLoc)
		node_client.Write([]byte("sgs" + strconv.Itoa(pi.identifier) + string(initial_game)))

	}
}

func ServerRegister(localIP string) shared.GameConfig {
	// Connect to server with RPC, port is always :8081
	serverConn, err := rpc.Dial("tcp", ":8081")
	if err != nil {
		log.Println("Cannot dial server. Please ensure the server is running and try again.")
		os.Exit(1)
	}
	var response shared.GameConfig
	// Get IP from server
	err = serverConn.Call("GServer.Register", localIP, &response)
	if err != nil {
		panic(err)
	}

	return response
}
