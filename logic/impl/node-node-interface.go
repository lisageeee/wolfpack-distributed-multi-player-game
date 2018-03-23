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
	"encoding/gob"
	"crypto/elliptic"
	"strconv"
)

// Node communication interface for communication with other player/logic nodes
type NodeCommInterface struct {
	PlayerNode			*PlayerNode
	PubKey 				*ecdsa.PublicKey
	PrivKey 			*ecdsa.PrivateKey
	Config 				shared.GameConfig
	ServerConn 			*rpc.Client
	IncomingMessages 	*net.UDPConn
	LocalAddr			net.Addr
	OtherNodes 			map[string]*net.UDPConn
	Connections 		[]string
}

type PlayerInfo struct {
	Address 			net.Addr
	PubKey 				ecdsa.PublicKey
}

// The message struct that is sent for all node communcation
type NodeMessage struct {
	Identifier string // the id of the sending node
	MessageType	string // identifies the type of message, can be: "move", "gameState", "connect", "connected"
	GameState * shared.GameState // a gamestate, included if MessageType is "gameState", else nil
	Move  * shared.Coord // a move, included if the message type is move
	Addr net.Addr // the address to connect to this node over, also an Identifier
}

// Creates a node comm interface with initial empty arrays
func CreateNodeCommInterface(pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey) (NodeCommInterface) {
	return NodeCommInterface {
		PubKey: pubKey,
		PrivKey: privKey,
		OtherNodes: make(map[string]*net.UDPConn),
		Connections: make([]string, 0)}
}

// Runs listener for messages from other nodes, should be run in a goroutine
func (n *NodeCommInterface) RunListener(listener *net.UDPConn, nodeListenerAddr string) {
	// Start the listener
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
			fmt.Println(string(buf[4:rlen]))
			err2 := json.Unmarshal(buf[4:rlen], &remoteCoord)
			if err2 != nil {
				panic(err2)
			}
			n.PlayerNode.GameState.PlayerLocs[id] = remoteCoord
			fmt.Println(n.PlayerNode.GameState.PlayerLocs)
		} else if string(buf[0:3]) == "sms"{
			var recState shared.GameRenderState
			err := json.Unmarshal(buf[4:rlen], &recState)
			if err != nil {
				panic(err)
			}
			id := string(buf[3])
			n.PlayerNode.GameState.PlayerLocs[id] = recState.PlayerLoc
			fmt.Println(recState.PlayerLoc)
			// TODO: Update go render state once other commit is merged
		}else if string(buf[0:rlen]) != "connected" {
			remoteClient, err := net.Dial("udp", string(buf[0:rlen]))
			if err != nil {
				panic(err)
			}
			toSend, err := json.Marshal(n.PlayerNode.GameState.PlayerLocs[n.PlayerNode.Identifier])
			// Code sgs sends the connecting node the position
			remoteClient.Write([]byte("sgs" + n.PlayerNode.Identifier + string(toSend)))
			n.OtherNodes[string(buf[0:rlen])] = remoteClient.(*net.UDPConn)
			// Keeping connections as array because we'll eventually get rid of it
			n.Connections = append(n.Connections, string(buf[0:rlen]))
		}
	}
}

// Registers the node with the server, receiving the gameconfig (and connections)
// TODO: maybe move this into node.go?
func (n *NodeCommInterface) ServerRegister() (id string) {
	gob.Register(&net.UDPAddr{})
	gob.Register(&elliptic.CurveParams{})
	gob.Register(&PlayerInfo{})

	if n.ServerConn == nil {
		// fmt.Printf("DEBUG - ServerRegister() n.ServerConn [%s] should be nil\n", n.ServerConn)
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
		playerInfo := PlayerInfo{n.LocalAddr, *n.PubKey}
		// fmt.Printf("DEBUG - PlayerInfo Struct [%v]\n", playerInfo)
		err = serverConn.Call("GServer.Register", playerInfo, &response)
		if err != nil {
			log.Fatal(err)
		}
		n.Config = response
	}

	n.GetNodes()

	// Start communcation with the other nodes
	go n.FloodNodes()

	return strconv.Itoa(n.Config.Identifier)
}

func (n *NodeCommInterface) GetNodes() {
	var response []net.Addr
	err := n.ServerConn.Call("GServer.GetNodes", *n.PubKey, &response)
	if err != nil {
		log.Fatal(err)
	}
	n.Connections = nil
	for _, addr := range response {
		n.Connections = append(n.Connections, addr.String())
	}
}

func (n *NodeCommInterface) SendHeartbeat() {
	for {
		var _ignored bool
		err := n.ServerConn.Call("GServer.Heartbeat", *n.PubKey, &_ignored)
		if err != nil {
			fmt.Printf("DEBUG - Heartbeat err: [%s]\n", err)
			n.ServerRegister()
		}
		boop := n.Config.GlobalServerHB
		time.Sleep(time.Duration(boop)*time.Microsecond)
	}
}
func(n* NodeCommInterface) SendMoveToNodes(move *shared.Coord){

	message := NodeMessage{
		MessageType: "move",
		Identifier: n.PlayerNode.Identifier,
		Move: move,
		Addr: n.LocalAddr,
		}

	toSend, _:= json.Marshal(message)
	for _, val := range n.OtherNodes{
		_, err := val.Write([]byte(toSend))
		if err != nil{
			fmt.Println(err)
		}
	}
}

func (n* NodeCommInterface) SendGameStateToNode(otherNodeId string){
	message := NodeMessage{
		MessageType: "move",
		Identifier: n.PlayerNode.Identifier,
		GameState: &n.PlayerNode.GameState,
		Addr: n.LocalAddr,
	}

	toSend, _:= json.Marshal(message)
	for _, val := range n.OtherNodes{
		_, err := val.Write([]byte(toSend))
		if err != nil{
			fmt.Println(err)
		}
	}
}

// Initiates connection with n.connections (provided nodes from server) on game init
func (n *  NodeCommInterface) FloodNodes() {
	const udpGeneric = "127.0.0.1:0"
	localIP, _ := net.ResolveUDPAddr("udp", udpGeneric)
	for _, ip := range n.Connections {
		nodeUdp, _ := net.ResolveUDPAddr("udp", ip)
		// Connect to other node
		nodeClient, err := net.DialUDP("udp", localIP, nodeUdp)
		n.OtherNodes[ip] = nodeClient
		if err != nil {
			panic(err)
		}
		// Exchange messages with other node
		myListener := n.LocalAddr.String()
		nodeClient.Write([]byte(myListener))
	}
}