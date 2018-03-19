package impl

import (
	"fmt"
	"net"
	"net/rpc"
	"log"
	"os"
	"../../shared"
	"strconv"
)

type NodeCommInterface struct {
	IncomingMessages *net.UDPConn
	Address *net.UDPAddr
	otherNodes []*net.UDPConn
	connections []string
}

func CreateNodeCommInterface () (NodeCommInterface) {
	return NodeCommInterface{}
}

func (n * NodeCommInterface) RunListener(nodeListenerAddr string) {
	// Start the listener
	addr, listener := StartListenerUDP(nodeListenerAddr)
	n.IncomingMessages = listener
	n.Address = addr
	listener.SetReadBuffer(1048576)

	i := 0
	for {
		i++
		buf := make([]byte, 1024)
		rlen, addr, err := listener.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(buf[0:rlen]))
		fmt.Println(addr)
		fmt.Println(i)
		// TODO: do whatever we do with the other node's messages

		//if string(buf[0:rlen]) != "connected" {
		//	remote_client, err := net.Dial("udp", string(buf[0:rlen]))
		//	if err != nil {
		//		panic(err)
		//	}
		//	remote_client.Write([]byte("connected"))
		//
		//	clients = append(clients, &remote_client)
		//}
	}
}

func (n * NodeCommInterface) ServerRegister() (shared.GameConfig) {
	// Connect to server with RPC, port is always :8081
	serverConn, err := rpc.Dial("tcp", ":8081")
	if err != nil {
		log.Println("Cannot dial server. Please ensure the server is running and try again.")
		os.Exit(1)
	}
	var response shared.GameConfig
	// Get IP from server
	err = serverConn.Call("GServer.Register", n.Address.String(), &response)
	if err != nil {
		panic(err)
	}
	n.connections = response.Connections
	return response
}

func (n *  NodeCommInterface) FloodNodes() {
	const udpGeneric = "127.0.0.1:0"
	localIP, _ := net.ResolveUDPAddr("udp", udpGeneric)
	for _, ip := range n.connections {
		nodeUdp, _ := net.ResolveUDPAddr("udp", ip)
		// Connect to other node
		nodeClient, err := net.DialUDP("udp", localIP, nodeUdp)
		if err != nil {
			panic(err)
		}
		// Exchange messages with other node
		myListener := n.Address.IP.String() + ":" +  strconv.Itoa(n.Address.Port)
		nodeClient.Write([]byte(myListener))
	}
}