package impl

import (
	"fmt"
	"net"
	"net/rpc"
	"log"
	"os"
	"../../shared"
	"encoding/json"
	"crypto/ecdsa"
	"time"
)

type NodeCommInterface struct {
	PlayerNode *PlayerNode
	PlayerInfo PlayerInfo
	PubKey ecdsa.PublicKey
	PrivKey ecdsa.PrivateKey
	Config shared.GameConfig
	ServerConn *rpc.Client
	IncomingMessages *net.UDPConn
	otherNodes []*net.Conn
	connections []string
}

type PlayerInfo struct {
	Address net.Addr
	PubKey ecdsa.PublicKey // TODO: who should be generating this?
}

func CreateNodeCommInterface() (NodeCommInterface) {
	return NodeCommInterface {
		otherNodes: make([]*net.Conn, 0),
		connections: make([]string, 0)}
}

func (n *NodeCommInterface) RunListener(nodeListenerAddr string) {
	// Start the listener
	addr, listener := StartListenerUDP(nodeListenerAddr)
	n.IncomingMessages = listener
	n.PlayerInfo.Address = addr
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
			remoteClient.Write([]byte("sgs" + n.PlayerNode.identifier + string(toSend)))
			n.otherNodes = append(n.otherNodes, &remoteClient)
		}
	}
}

func (n *NodeCommInterface) ServerRegister() (pubKeyStr string) {
	// Connect to server with RPC, port is always :8081
	serverConn, err := rpc.Dial("tcp", ":8081")
	if err != nil {
		log.Println("Cannot dial server. Please ensure the server is running and try again.")
		os.Exit(1)
	}
	// Storing in object so that we can do other RPC calls outside of this function
	n.ServerConn = serverConn

	var response shared.GameConfig
	// Register with server
	err = serverConn.Call("GServer.Register", n.PlayerInfo, &response)
	if err != nil {
		log.Fatal(err)
	}
	n.Config = response

	n.GetNodes()

	return "hah" // TODO: pubkey
}

func (n *NodeCommInterface) GetNodes() {
	response := make([]net.Addr, 0)
	err := n.ServerConn.Call("GServer.GetNodes", n.PubKey, &response)
	if err != nil {
		log.Fatal(err)
	}

	for _, addr := range response {
		n.connections = append(n.connections, addr.String())
	}
}

func (n *NodeCommInterface) SendHeartbeat() {
	for {
		var _ignored bool
		err := n.ServerConn.Call("RServer.Heartbeat", n.PubKey, &_ignored)
		if err != nil {
			n.ServerRegister()
		}
		boop := n.Config.GlobalServerHB
		time.Sleep(time.Duration(boop)*time.Microsecond)
	}
}

func (n *NodeCommInterface) FloodNodes() {
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
		myListener := n.PlayerInfo.Address.String()
		nodeClient.Write([]byte(myListener))
	}
}