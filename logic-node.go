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
	"./shared"
	"./geometry"
	"encoding/json"
)

type RemotePlayerInterface struct {
	pixelListener     *net.UDPConn
	pixelWriter 	  *net.UDPConn
	playerCommChannel chan string
	otherNodes        *net.UDPConn
	playerPosition    shared.Coord
	step              int
	geo               geometry.GridManager
}

// Entrypoint, sets up communication channels and creates the RemotePlayerInterface
func main() {
	fmt.Println("hello world")

	// Listener IP address
	var node_listener_ip_address string
	var player_listener_ip_address string
	var pixel_ip_address string
	// Can start with an IP as param
	if len(os.Args)>2 {
		node_listener_ip_address = os.Args[1]
		player_listener_ip_address = os.Args[2]
	}else{
		node_listener_ip_address = "127.0.0.1:0"
		player_listener_ip_address = "127.0.0.1:12345"
		pixel_ip_address = "127.0.0.1:1234"
	}

	_, player := startListener(player_listener_ip_address)
	defer player.Close()

	// Start the node to node interface
	_, client := startListener(node_listener_ip_address)
	defer client.Close()

	// Start the pixel interface
	pixel := setupUDPToPixel(pixel_ip_address)
	defer pixel.Close()

	otherNodes := serverRegister(client.LocalAddr().String())
	udpAddr := client.LocalAddr().(*net.UDPAddr)
	floodNodes(otherNodes, udpAddr)

	pi := RemotePlayerInterface{pixelListener: player, pixelWriter: pixel, otherNodes: client, playerCommChannel: make(chan string),
	geo: geometry.CreateNewGridManager(20, 20, []shared.Coord{{4, 3}})}
	pi.runGame()
}

func (pi * RemotePlayerInterface) runGame() {
	go pi.runNodeListener()
	go pi.runPlayerListener()

	for {
		message := <-pi.playerCommChannel
		switch message {
		case "quit":
			break
		default:
			pi.movePlayer(message)
			pi.sendPlayerGameState()
			fmt.Println("movin' player", message)
		}
	}

}

func (pi * RemotePlayerInterface) movePlayer(move string) {
	newPosition := shared.Coord{X: pi.playerPosition.X, Y: pi.playerPosition.Y}
	switch move {
		case "up":
			newPosition.Y = newPosition.Y + pi.step
		case "down":
			newPosition.Y = newPosition.Y - pi.step
		case "left":
			newPosition.X = newPosition.X - pi.step
		case "right":
			newPosition.X = newPosition.X + pi.step
	}
	if pi.geo.IsValidMove(newPosition) {
		pi.playerPosition = newPosition
	}
}

func (pi *RemotePlayerInterface) sendPlayerGameState() {
	// TODO: add other nodes
	toSend, err := json.Marshal(pi.playerPosition)
	if err != nil {
		fmt.Println(err)
	} else {
		// Send position to player node
		pi.pixelWriter.Write([]byte(toSend))
	}
}

func startListener(ip_addr string) (*net.UDPAddr, *net.UDPConn) {
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

func (pi * RemotePlayerInterface) runPlayerListener() {
	// takes a listener client
	// runs the listener in a infinite loop
	player := pi.pixelListener
	player.SetReadBuffer(1048576)

	i := 0
	for {
		i++
		buf := make([]byte, 1024)
		rlen, _, err := player.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		}
		pi.playerCommChannel <- string(buf[0:rlen])
	}
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
		if string(buf[0:rlen]) != "connected" {
			remote_client, err := net.Dial("udp", string(buf[0:rlen]))
			if err != nil {
				panic(err)
			}
			remote_client.Write([]byte("connected"))

			clients = append(clients, &remote_client)
		}
	}
}

func floodNodes(otherNodes []string, udp_addr *net.UDPAddr) {
	localIP, _ := net.ResolveUDPAddr("udp", udp_generic)
	for _, ip := range otherNodes {
		node_udp, _ := net.ResolveUDPAddr("udp", ip)
		// Connect to other node
		node_client, err := net.DialUDP("udp", localIP, node_udp)
		if err != nil {
			panic(err)
		}
		// Exchange messages with other node
		myListener := udp_addr.IP.String() + ":" +  strconv.Itoa(udp_addr.Port)
		node_client.Write([]byte(myListener))
	}
}

func serverRegister(localIP string) []string {
	// Connect to server with RPC, port is always :8081
	serverConn, err := rpc.Dial("tcp", ":8081")
	if err != nil {
		log.Println("Cannot dial server. Please ensure the server is running and try again.")
		os.Exit(1)
	}
	var response []string
	// Get IP from server
	err = serverConn.Call("GServer.Register", localIP, &response)
	if err != nil {
		panic(err)
	}
	if len(response) > 0 {
		for ind, val := range response {
			fmt.Println(strconv.Itoa(ind) + ": " + val)
		}
	}
	return response
}

func setupUDPToPixel(ip_addr string) (*net.UDPConn) {
	node_udp, _ := net.ResolveUDPAddr("udp", ip_addr)
	node_client, err := net.DialUDP("udp", nil, node_udp)
	if err != nil {
		panic(err)
	}
	return node_client
}