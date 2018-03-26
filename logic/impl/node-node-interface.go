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
	ServerAddr			string
	ServerConn 			*rpc.Client
	IncomingMessages 	*net.UDPConn
	LocalAddr			net.Addr
	OtherNodes 			map[string]*net.UDPConn
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
	Addr string // the address to connect to this node over
}

// Creates a node comm interface with initial empty arrays
func CreateNodeCommInterface(pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey, serverAddr string) (NodeCommInterface) {
	return NodeCommInterface {
		PubKey: pubKey,
		PrivKey: privKey,
		ServerAddr : serverAddr,
		OtherNodes: make(map[string]*net.UDPConn),
		}
}

// Runs listener for messages from other nodes, should be run in a goroutine
func (n *NodeCommInterface) RunListener(listener *net.UDPConn, nodeListenerAddr string) {
	// Start the listener
	listener.SetReadBuffer(1048576)

	i := 0
	var message NodeMessage
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

		json.Unmarshal(buf[0:rlen], &message)

		switch message.MessageType {
			case "gameState":
				n.HandleReceivedGameState(message.Identifier, message.GameState)
			case "move":
				n.HandleReceivedMove(message.Identifier, message.Move)
			case "connect":
				n.HandleIncomingConnectionRequest(message.Identifier, message.Addr)
			case "connected":
			// Do nothing
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
		serverConn, err := rpc.Dial("tcp", n.ServerAddr)
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
	var response map[string]net.Addr
	err := n.ServerConn.Call("GServer.GetNodes", *n.PubKey, &response)
	if err != nil {
		log.Fatal(err)
	}

	n.OtherNodes = make(map[string]*net.UDPConn)

	for id, addr := range response {
		nodeClient := n.GetClientFromAddrString(addr.String())
		nodeUdp, _ := net.ResolveUDPAddr("udp", addr.String())
		// Connect to other node
		nodeClient, err := net.DialUDP("udp", nil, nodeUdp)
		if err != nil {
			panic(err)
		}
		n.OtherNodes[id] = nodeClient
	}
}

func (n *NodeCommInterface) GetClientFromAddrString(addr string) (*net.UDPConn) {
	nodeUdp, _ := net.ResolveUDPAddr("udp", addr)
	// Connect to other node
	nodeClient, err := net.DialUDP("udp", nil, nodeUdp)
	if err != nil {
		panic(err)
	}
	return nodeClient
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

	if move == nil {
		return
	}

	message := NodeMessage{
		MessageType: "move",
		Identifier: n.PlayerNode.Identifier,
		Move: move,
		Addr: n.LocalAddr.String(),
		}

	toSend, err := json.Marshal(&message)

	if err != nil {
		fmt.Println(err)
	}
	for _, val := range n.OtherNodes{
		_, err := val.Write(toSend)
		if err != nil{
			fmt.Println(err)
		}
	}
}

func (n* NodeCommInterface) SendGameStateToNode(otherNodeId string){
	message := NodeMessage{
		MessageType: "gameState",
		Identifier: n.PlayerNode.Identifier,
		GameState: &n.PlayerNode.GameState,
		Addr: n.LocalAddr.String(),
	}

	toSend, _:= json.Marshal(&message)
	n.OtherNodes[otherNodeId].Write(toSend)
}

func (n* NodeCommInterface) HandleReceivedGameState(identifier string, gameState *shared.GameState) {
	//TODO: don't just wholesale replace this
	n.PlayerNode.GameState = *gameState
}

func (n* NodeCommInterface) HandleReceivedMove(identifier string, move *shared.Coord) {
	// TODO: add checks
	// Need nil check for bad move
	if move != nil {
		n.PlayerNode.GameState.PlayerLocs[identifier] = *move
	}
}

func (n* NodeCommInterface) HandleIncomingConnectionRequest(identifier string, addr string) {
	node := n.GetClientFromAddrString(addr)
	n.OtherNodes[identifier] = node
}

func (n* NodeCommInterface) InitiateConnection(nodeClient *net.UDPConn) {
	message := NodeMessage{
		MessageType: "connect",
		Identifier: strconv.Itoa(n.Config.Identifier),
		GameState: nil,
		Addr: n.LocalAddr.String(),
		Move: nil,
	}

	toSend, _:= json.Marshal(&message)
	for _, val := range n.OtherNodes{
		_, err := val.Write(toSend)
		if err != nil{
			fmt.Println(err)
		}
	}
}

// Sends connection message to connections after receiving from server
func (n *  NodeCommInterface) FloodNodes() {
	for _, nodeClient := range n.OtherNodes {
		// Exchange messages with other node
		n.InitiateConnection(nodeClient)
	}
}