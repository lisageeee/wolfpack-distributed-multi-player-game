package impl

import (
	"fmt"
	"net"
	"net/rpc"
	"log"
	"os"
	"../../shared"
	"strconv"
	"encoding/json"
)

type NodeCommInterface struct {
	PlayerNode *PlayerNode
	IncomingMessages *net.UDPConn
	Address *net.UDPAddr
	otherNodes []*net.Conn
	connections []string
}

func CreateNodeCommInterface () (NodeCommInterface) {
	return NodeCommInterface{otherNodes: make([]*net.Conn, 0), connections: make([]string, 0)}
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
		// Write to the node comm channel
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
			n.PlayerNode.GameRenderState.OtherPlayers[id] = remoteCoord
			fmt.Println(n.PlayerNode.GameRenderState.OtherPlayers)
		} else if string(buf[0:rlen]) != "connected" {
			remoteClient, err := net.Dial("udp", string(buf[0:rlen]))
			if err != nil {
				panic(err)
			}
			toSend, err := json.Marshal(n.PlayerNode.GameRenderState.PlayerLoc)
			// Code sgs sends the connecting node the gamestate
			remoteClient.Write([]byte("sgs" + strconv.Itoa(n.PlayerNode.identifier) + string(toSend)))
			n.otherNodes = append(n.otherNodes, &remoteClient)
		}
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