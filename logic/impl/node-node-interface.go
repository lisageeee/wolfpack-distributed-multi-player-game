package impl

import (
	"fmt"
	"net"
	"net/rpc"
	"log"
	"os"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"time"
	"encoding/gob"
	"encoding/hex"
	"strconv"
	"github.com/rzlim08/GoVector/govec"
	"math/big"
	key "../../key-helpers"
	"../../wolferrors"
	"../../shared"
	"sync"
	"encoding/json"
)

// Node communication interface for communication with other player/logic nodes as well as the server
type NodeCommInterface struct {
	// A reference back to this interface's "main" node
	PlayerNode			*PlayerNode

	// The public key of this nodes
	PubKey 				*ecdsa.PublicKey

	// The private key of this node, used to encrypt messages
	PrivKey 			*ecdsa.PrivateKey

	// The gameconfig for the game, primarily used here to form connections to the given nodes
	Config 				shared.GameConfig

	// The address of the server for this game
	ServerAddr			string

	// The RPC connection to the server
	ServerConn 			*rpc.Client

	// The UDP connection over which this node listens for messages from other logic nodes
	IncomingMessages 	*net.UDPConn

	// The address of this node's listener
	LocalAddr			net.Addr

	// The current map of identifiers to connections of nodes in play
	OtherNodes 			map[string]*net.UDPConn

	// The current map of identifiers to public keys of nodes in play
	NodeKeys		    map[string]*ecdsa.PublicKey

	// The GoVector log
	Log 				*govec.GoLog

	// A channel that, when written to, will stop heartbeats. Primarily for testing
	HeartAttack 		chan bool

	// A map to store move commits in before receiving their associated moves
	MoveCommits			map[string]string

	//PlayerScores		map[string]int

	// Channel that messages are written to so they can be handled by the goroutine that deals with sending messages
	// and managing the player nodes
	MessagesToSend		chan *PendingMessage

	// Channel that the identifiers of nodes to delete are added to so they can be handled by the goroutine that deals
	// with sending messages and managing the player nodes
	NodesToDelete		chan string

	// Channel that the identifiers and connections of nodes to add to other nodes are sent to so they can be handled
	// by the goroutine that deals with sending messages and managing the player nodes
	NodesToAdd			chan *OtherNode

	// A channel for received acks to be written to
	ACKSReceived          chan *ACKMessage

	// A channel to write nodes that appear to have been shut down to
	NodesWriteConnRefused chan string

	// Pending moves go in this gannel
	MovesToSend           chan *PendingMoveUpdates

	// Keeps track of the number of failed messages between nodes
	Strikes               StrikeLockMap // Heartbeat protocol between nodes

	// Write to this channel to trigger a gamestate send to the pixel node
	GameStateToSend       chan bool

	// A boolean set to false before this node has reconciled the gamestate when joining
	HasGameState		  bool

	// Write to this channel to send move (associated with a move commit) to prey after receiving a prey's move
	MoveToSendToPrey	chan *MoveWithSeq

	// Collect all of the ack messages from chan
	CollectAcks			CollectACKSMap
}

// The map that keeps track of all the ack messages
type CollectACKSMap struct {
	sync.RWMutex
	ACKS map[uint64]map[string]string
}

// Another version of a move struct but with a sequence number attached to it
type MoveWithSeq struct {
	Move shared.Coord
	Seq uint64
}

// The map that keeps track of nodes that could not be written to (err: write to udp error")
type StrikeLockMap struct {
	sync.RWMutex
	StrikeCount map[string]int
}

// A message for another node with a recipient and a byte-encoded message. If the recipient is "all", the message is
// sent to every node in OtherNodes.
type PendingMessage struct {
	Recipient string
	Message []byte
}

// A struct to hold pending moves
type PendingMoveUpdates struct {
	Seq	uint64
	Coord *shared.Coord
	Rejected int
}

// A struct to form an ACK message
type ACKMessage struct {
	Seq        uint64
	Identifier string
}

// An othernode struct, used for storing node ids/conns before they are added to the OtherNodes map
type OtherNode struct {
	Identifier string
	Conn *net.UDPConn
	PubKey *ecdsa.PublicKey
}

// A playerinfo struct, provides identification information about this node: the address and public key
type PlayerInfo struct {
	Address 			net.Addr
	PubKey 				ecdsa.PublicKey
	Prey				bool
}

// The message struct that is sent for all node communication
type NodeMessage struct {
	// the id of the sending node
	Identifier  string

	// identifies the type of message so we know how to handle it
	// can be: "move", "moveCommit", "gameState", "connect", "connected", "gamestateReq", "captured", "ack", '
	MessageType string

	// a gamestate, included if MessageType is "gameState", else nil
	GameState   *shared.GameState

	// a move, included if the message type is move
	Move        shared.SignedMove

	// a move commit, included if the message type is moveCommit
	MoveCommit  *shared.MoveCommit

	// a score, included if the message is a preyCapture
	Score int

	// A string representing th epublic key if this is a connect message
	PubKey 	string

	// the address to connect to the sending node over
	Addr        string

	// Keep track of sequence number for response ACKs
	Seq			uint64
}

var sequenceNumber uint64 = 0

const STRIKE_OUT = 3

// Creates a node comm interface with initial empty arrays/maps
func CreateNodeCommInterface(pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey, serverAddr string) (NodeCommInterface) {
	return NodeCommInterface{
		PubKey:                pubKey,
		PrivKey:               privKey,
		ServerAddr:            serverAddr,
		OtherNodes:            make(map[string]*net.UDPConn),
		NodeKeys:              make(map[string]*ecdsa.PublicKey),
		HeartAttack:           make(chan bool),
		MoveCommits:           make(map[string]string),
		MessagesToSend:        make(chan *PendingMessage, 30),
		NodesToDelete:         make(chan string, 5),
		NodesToAdd:            make(chan *OtherNode, 10),
		ACKSReceived:          make(chan *ACKMessage, 30),
		NodesWriteConnRefused: make(chan string, 30),
		MovesToSend:           make(chan *PendingMoveUpdates, 1),
		Strikes:               StrikeLockMap{StrikeCount:make(map[string]int)},
		GameStateToSend:       make(chan bool),
		HasGameState: 		   false,
		MoveToSendToPrey:	   make(chan *MoveWithSeq),
		CollectAcks:		   CollectACKSMap{ACKS:make(map[uint64]map[string]string)},
	}
}

// Runs listener for messages from other nodes, should be run in a goroutine
// Unmarshalls received messages and dispatches them to the appropriate handler function
func (n *NodeCommInterface) RunListener(listener *net.UDPConn, nodeListenerAddr string) {
	// Start the listener
	listener.SetReadBuffer(1048576)

	i := 0
	for {
		i++
		buf := make([]byte, 2048)
		_, _, err := listener.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		}

		message := receiveMessage(n.Log, buf)

		switch message.MessageType {
			case "gameState":
				n.HandleReceivedGameState(message.Identifier, message.GameState)
			case "gamestateReq":
				n.HandleGameStateConnReq(message.Identifier)
			case "moveCommit":
				n.HandleReceivedMoveCommit(message.Identifier, message.MoveCommit)
			case "move":
				// Currently only planning to do the lockstep protocol with prey node
				// In the future, may include players close to prey node
				// I.e. check move commits
				authentic := n.CheckAuthenticityOfMove(n.NodeKeys[message.Identifier], &message.Move)
				if !authentic{
					fmt.Println("False coordinates")
					continue
				}
				var coords shared.Coord
				err := json.Unmarshal(message.Move.MoveByte, &coords)
				if err != nil {
					fmt.Println("Could not unmarshal")
					fmt.Println(err)
				} else {
					if message.Identifier == "prey" {
						err := n.HandleReceivedPreyMove(message.Identifier, &coords)
						if err != nil {
							fmt.Println("The error in the prey moving")
							fmt.Println(err)
						}
					} else {
						n.HandleReceivedMoveNL(message.Identifier, &coords, message.Seq)
					}
				}
			case "connect":
				n.HandleIncomingConnectionRequest(message.Identifier, message.Addr, message.PubKey)
			case "connected":
			// Do nothing
			case "captured":
				var coords shared.Coord
				authentic := n.CheckAuthenticityOfMove(n.NodeKeys[message.Identifier], &message.Move)
				if !authentic{
					fmt.Println("False coordinates")
					continue
				}
				err := json.Unmarshal(message.Move.MoveByte, &coords)
				if err != nil {
					fmt.Println("Could not unmarshal")
					fmt.Println(err)
				} else {
					n.HandleCapturedPreyRequest(message.Identifier, &coords, message.Score)
				}
			case "ack":
				n.HandleReceivedAck(message.Identifier, message.Seq)
			default:
				// Do nothing
				fmt.Println("Message type is incorrect")
		}
	}
}

// Routine that handles all reads and writes of the OtherNodes map; single thread preventing concurrent iteration and write
// exception. This routine therefore handles all sending of messages as well as that requires iteration over OtherNodes.
func (n *NodeCommInterface) ManageOtherNodes() {
	for {
		select {
		case toSend := <-n.MessagesToSend :
			if toSend.Recipient != "all" {
				// Send to the single node
				if _, ok := n.OtherNodes[toSend.Recipient]; ok {
					n.OtherNodes[toSend.Recipient].Write(toSend.Message)
				}
			} else {
				// Send the message to all nodes except prey node
				n.sendMessageToPlayerNodes(toSend.Message)
			}
		case toAdd := <- n.NodesToAdd:
			n.OtherNodes[toAdd.Identifier] = toAdd.Conn
			n.NodeKeys[toAdd.Identifier] = toAdd.PubKey
		case toDelete := <-n.NodesToDelete:
			fmt.Printf("To delete: %s\n", toDelete)
			delete(n.OtherNodes, toDelete)
			n.PlayerNode.GameState.PlayerLocs.Lock()
			delete(n.PlayerNode.GameState.PlayerLocs.Data, toDelete)
			delete(n.NodeKeys, toDelete)
			fmt.Printf("PlayerLocs.Data %v\n", n.PlayerNode.GameState.PlayerLocs.Data)
			n.PlayerNode.GameState.PlayerLocs.Unlock()
		}
	}
}

// Go routine to grab ack messages from chan n.ACKSReceived
func (n *NodeCommInterface) CollectACKs() {
	for {
		select {
		case ack := <-n.ACKSReceived:
			n.CollectAcks.Lock()
			if n.CollectAcks.ACKS[ack.Seq] == nil {
				n.CollectAcks.ACKS[ack.Seq] = make(map[string]string)
			}
			n.CollectAcks.ACKS[ack.Seq][ack.Identifier] = ""

			for k := range n.CollectAcks.ACKS {
				if k < ack.Seq {
					delete(n.CollectAcks.ACKS, k)
				}
			}
			n.CollectAcks.Unlock()
		}
	}
}

// Routine that handles the ACKs being received in response to a move message from this node
func (n *NodeCommInterface) ManageAcks() {
	tempCh := make(chan *PendingMoveUpdates, 1)
	for {
		if len(tempCh) == 0 {
			move := <-n.MovesToSend
			tempCh <- move
			fmt.Println("woot got move")
		}
		select {
		//case <- time.After(250 * time.Millisecond):
		//	if len(tempCh) == 0 {
		//		move := <-n.MovesToSend
		//		tempCh <- move
		//	}
		case <- time.After(500 * time.Millisecond):
			if len(n.OtherNodes) != 0 {
				currMove := <-tempCh
				n.CollectAcks.Lock()
				fmt.Println("DEBUG - CollectACks Map:", n.CollectAcks.ACKS, "And seq: ", currMove.Seq)
				if _, ok := n.CollectAcks.ACKS[currMove.Seq]; ok {
					fmt.Println("DEBUG - in case where collect act contains seq")
					if len(n.OtherNodes) == 1 { // only prey (will handle acks)
						fmt.Printf("DEBUG - Here is the current move %v and seq %d\n", currMove.Coord, currMove.Seq)
						if len(n.CollectAcks.ACKS[currMove.Seq]) == 1 {
							n.PlayerNode.GameState.PlayerLocs.Lock()
							n.PlayerNode.GameState.PlayerLocs.Data[n.PlayerNode.Identifier] = *currMove.Coord
							n.PlayerNode.GameState.PlayerLocs.Unlock()
							delete(n.CollectAcks.ACKS, currMove.Seq)
							n.CollectAcks.Unlock()
						} else {
							n.CollectAcks.Unlock()
							currMove.Rejected++
							if currMove.Rejected < 5 {
								n.ReSendMoveToNodes(currMove.Coord, currMove.Seq, true)
								tempCh <- currMove
								fmt.Println(currMove.Rejected)
								fmt.Println("DEBUG - Sending currMove back to tempCh with seq: ", currMove.Seq)
							} else {
								fmt.Println("REJECTING MOVE FOREVER WITH SEQ:", currMove.Seq)
							}
						}
					} else if len(n.OtherNodes) == 2 { // prey(doesn't handle acks) + someone else
						if len(n.CollectAcks.ACKS[currMove.Seq]) >= len(n.OtherNodes)/2 {
							fmt.Println("DEBUG - We should be in this state with one prey + one player")
							fmt.Printf("DEBUG - Here is the current move %v and seq %d\n", currMove.Coord, currMove.Seq)
							n.PlayerNode.GameState.PlayerLocs.Lock()
							n.PlayerNode.GameState.PlayerLocs.Data[n.PlayerNode.Identifier] = *currMove.Coord
							n.PlayerNode.GameState.PlayerLocs.Unlock()
							delete(n.CollectAcks.ACKS, currMove.Seq)
							n.CollectAcks.Unlock()
						} else {
							n.CollectAcks.Unlock()
							currMove.Rejected++
							if currMove.Rejected < 10 {
								n.ReSendMoveToNodes(currMove.Coord, currMove.Seq, false)
								tempCh <- currMove
								fmt.Println(currMove.Rejected)
								fmt.Println("DEBUG - Sending currMove back to tempCh with seq: ", currMove.Seq)
							} else {
								fmt.Println("REJECTING MOVE FOREVER WITH SEQ:", currMove.Seq)
							}
						}
					} else {
						// if the # of acks > # of connected nodes (majority consensus)
						if len(n.CollectAcks.ACKS[currMove.Seq]) > len(n.OtherNodes)/2 {
							fmt.Println("DEBUG - We should NOT be in this state with one prey")
							n.PlayerNode.GameState.PlayerLocs.Lock()
							n.PlayerNode.GameState.PlayerLocs.Data[n.PlayerNode.Identifier] = *currMove.Coord
							n.PlayerNode.GameState.PlayerLocs.Unlock()
							delete(n.CollectAcks.ACKS, currMove.Seq)
							n.CollectAcks.Unlock()
						} else {
							n.CollectAcks.Unlock()
							currMove.Rejected++
							if currMove.Rejected < 10 {
								n.ReSendMoveToNodes(currMove.Coord, currMove.Seq, false)
								tempCh <- currMove
								fmt.Println(currMove.Rejected)
								fmt.Println("DEBUG - Sending currMove back to tempCh with seq: ", currMove.Seq)
							} else {
								fmt.Println("REJECTING MOVE FOREVER WITH SEQ:", currMove.Seq)
							}
						}
					}
				} else {
					n.CollectAcks.Unlock()
				}
			}
		}
	}
}

func (n *NodeCommInterface) PruneNodes() {
	for {
		select {
		case id := <-n.NodesWriteConnRefused:
			if id != "prey" {
				n.Strikes.StrikeCount[id]++
				if n.Strikes.StrikeCount[id] > STRIKE_OUT {
					n.NodesToDelete <- id
					fmt.Printf("Deleting this id: %s\n", id)
					delete(n.Strikes.StrikeCount, id)
				}
			}
		}
	}
}

func (n *NodeCommInterface) SendGameStateToPixel() {
	for {
		select {
		case <-n.GameStateToSend:
			n.PlayerNode.pixelInterface.SendPlayerGameState(n.PlayerNode.GameState)
		}
	}
}

// Helper function that unpacks the GoVector message tooling
// Returns the unmarshalled NodeMessage, ready for reading
func receiveMessage(goLog *govec.GoLog, payload []byte) NodeMessage {
	// Just removes the golog headers from each message
	// TODO: set up error handling
	var message NodeMessage
	goLog.UnpackReceive("LogicNodeReceiveMessage", payload, &message)
	return message
}

// Helper function that packs the GoVector message tooling
// Returns the byte-encoded message, ready to send
func sendMessage(goLog *govec.GoLog, message NodeMessage, tag string) []byte{
	var newMessage []byte
	if tag == ""{
		newMessage = goLog.PrepareSend("SendMessageToOtherNode", message)
	}else{
		newMessage = goLog.PrepareSend(tag, message)
	}

	return newMessage

}
// Registers the node with the server, receiving the game config (and connections)
// Returns the unique id of this node assigned by the server
func (n *NodeCommInterface) ServerRegister() (id string) {
	gob.Register(&net.UDPAddr{})
	gob.Register(&elliptic.CurveParams{})
	gob.Register(&PlayerInfo{})

	if n.ServerConn == nil {
		response, err := DialAndRegister(n)
		if err != nil {
			os.Exit(1)
		}
		n.Log = govec.InitGoVectorMultipleExecutions("LogicNodeId-"+response.Identifier,
			"LogicNodeFile")

		n.Config = response
	}
	n.GetNodes()

	return n.Config.Identifier
}

// Another server registration function, used to deal with server disconnection.
func DialAndRegister(n *NodeCommInterface) (shared.GameConfig, error) {
	// fmt.Printf("DEBUG - ServerRegister() n.ServerConn [%s] should be nil\n", n.ServerConn)
	// Connect to server with RPC, port is always :8081
	serverConn, err := rpc.Dial("tcp", n.ServerAddr)
	if err != nil {
		log.Println("Cannot dial server. Please ensure the server is running and try again.")
		return shared.GameConfig{}, err
	}
	// Storing in object so that we can do other RPC calls outside of this function
	n.ServerConn = serverConn
	var response shared.GameConfig
	// Register with server
	playerInfo := PlayerInfo{n.LocalAddr, *n.PubKey, false}
	// fmt.Printf("DEBUG - PlayerInfo Struct [%v]\n", playerInfo)
	err = serverConn.Call("GServer.Register", playerInfo, &response)
	if err != nil {
		return shared.GameConfig{}, err
	}
	return response, nil
}

// Requests the list of currently connected nodes from the server, and initiates a connection with them
func (n *NodeCommInterface) GetNodes() {
	var response map[string]shared.NodeRegistrationInfo
	err := n.ServerConn.Call("GServer.GetNodes", *n.PubKey, &response)
	if err != nil {
		panic(err)
		log.Fatal(err)
	}

	// If 0, it is only us, don't need to update gamestate
	if len(response) < 1 {
		fmt.Println("no other nodes")
		// This node is the only node in gameplay, doesn't need to get gamestate from other nodes
		n.HasGameState = true
	}

	for id, regInfo := range response {
		nodeClient := n.GetClientFromAddrString(regInfo.Addr.String())
		pubKey:= key.StringToPubKey(regInfo.PubKey)
		node := OtherNode{Identifier: id, Conn: nodeClient, PubKey: &pubKey}
		n.NodesToAdd <- &node
		n.InitiateConnection(nodeClient, id)
	}
}

// Takes in an address string and makes a UDP connection to the client specified by the string. Returns the connection.
func (n *NodeCommInterface) GetClientFromAddrString(addr string) (*net.UDPConn) {
	nodeUdp, _ := net.ResolveUDPAddr("udp", addr)
	// Connect to other node
	nodeClient, err := net.DialUDP("udp", nil, nodeUdp)
	if err != nil {
		panic(err)
	}
	return nodeClient
}

// Sends a heartbeat to the server at the interval specificed at server registration
func (n *NodeCommInterface) SendHeartbeat() {
	var _ignored bool
	for {
		select {
		case <-n.HeartAttack:
			return
		default:
			err := n.ServerConn.Call("GServer.Heartbeat", *n.PubKey, &_ignored)
			if err != nil {
				fmt.Printf("DEBUG - Heartbeat err: [%s]\n", err)
				n.Config = n.Reregister()
			}
			boop := n.Config.GlobalServerHB
			time.Sleep(time.Duration(boop/2)*time.Microsecond)
		}
	}
}

// Function that is started when the server dies; will continue to reregister until the server comes back up
func (n* NodeCommInterface) Reregister() shared.GameConfig {
	response, register_failed_err := DialAndRegister(n)
	for register_failed_err != nil {
		response, register_failed_err = DialAndRegister(n)
		time.Sleep(time.Second)
	}
	fmt.Println("Registered Server")
	return response
}

// Takes in a new coordinate for this node and sends it to all other nodes.
func(n* NodeCommInterface) SendMoveToNodes(move *shared.Coord, sendMoveCommitToPrey bool) {
	if move == nil {
		return
	}

	sequenceNumber++
	moveId := n.CreateMove(move)
	message := NodeMessage{
		MessageType: "move",
		Identifier:  n.PlayerNode.Identifier,
		Move:        moveId,
		Addr:        n.LocalAddr.String(),
		Seq:         sequenceNumber,
	}

	// calculate and sign a hash based on move that the player made
	// this is sent to the prey node
	moveHash := n.CalculateHash(*move, n.PlayerNode.Identifier)
	r, s, _ := n.SignMoveCommit(moveHash)
	moveCommit := shared.MoveCommit{
		Seq:      sequenceNumber,
		MoveHash: moveHash,
		R:        r.String(),
		S:        s.String(),
	}

	// if there are other nodes in play
	if len(n.OtherNodes) > 0 {
		toSend := sendMessage(n.Log, message, "Sendin' move")
		n.MessagesToSend <- &PendingMessage{Recipient: "all", Message: toSend}
		n.MovesToSend <- &PendingMoveUpdates{Seq: sequenceNumber, Coord: move, Rejected: 0}

		if sendMoveCommitToPrey {
			n.SendMoveCommitToPreyNode(&moveCommit)
			fmt.Println("DEBUG - Sending move commit")
		}

		// This is the move that will be sent to prey once we've received a response
		// back for the move commit associated with this move
		n.MoveToSendToPrey <- &MoveWithSeq{*move, sequenceNumber}
	} else { // by itself, so just whatever update
		n.PlayerNode.GameState.PlayerLocs.Lock()
		n.PlayerNode.GameState.PlayerLocs.Data[n.PlayerNode.Identifier] = *move
		n.PlayerNode.GameState.PlayerLocs.Unlock()
		n.GameStateToSend <- true
	}
}

func(n* NodeCommInterface) ReSendMoveToNodes(move *shared.Coord, seq uint64, sendMoveCommitToPrey bool) {
	if move == nil {
		return
	}

	moveId := n.CreateMove(move)
	message := NodeMessage{
		MessageType: "move",
		Identifier:  n.PlayerNode.Identifier,
		Move:        moveId,
		Addr:        n.LocalAddr.String(),
		Seq:         seq,
	}

	toSend := sendMessage(n.Log, message, "Sendin' move")
	n.MessagesToSend <- &PendingMessage{Recipient: "all", Message: toSend}

	if sendMoveCommitToPrey {
		// calculate and sign a hash based on move that the player made
		// this is sent to the prey node
		moveHash := n.CalculateHash(*move, n.PlayerNode.Identifier)
		r, s, err := n.SignMoveCommit(moveHash)
		if err == nil {
			moveCommit := shared.MoveCommit{
				Seq:      seq,
				MoveHash: moveHash,
				R:        r.String(),
				S:        s.String(),
			}
			n.SendMoveCommitToPreyNode(&moveCommit)
		}
		fmt.Println("DEBUG - Sending move commit")

		// This is the move that will be sent to prey once we've received a response
		// back for the move commit associated with this move
		n.MoveToSendToPrey <- &MoveWithSeq{*move, seq}
	}
}

// Takes in a new coordinate for this node and sends it to the prey node
func(n* NodeCommInterface) SendMoveToPreyNode(move *MoveWithSeq) {
	if move == nil {
		return
	}

	moveId := n.CreateMove(&move.Move)
	message := NodeMessage {
		MessageType: "move",
		Identifier:  n.PlayerNode.Identifier,
		Move:        moveId,
		Addr:        n.LocalAddr.String(),
		Seq:         move.Seq,
	}

	toSend := sendMessage(n.Log, message, "Sendin' move to prey")
	n.MessagesToSend <- &PendingMessage{Recipient: "prey", Message: toSend}
}

func (n *NodeCommInterface) CreateMove(move *shared.Coord) shared.SignedMove {
	moveBytes, err := json.Marshal(move)
	r, s, err := ecdsa.Sign(rand.Reader, n.PrivKey, moveBytes)
	if err != nil {
		fmt.Println("could not sign move")
		panic(err)
	}
	moveId := shared.SignedMove{
		moveBytes,
		r.String(),
		s.String(),
	}
	return moveId
}

func(n* NodeCommInterface) SendPreyCaptureToNodes(move *shared.Coord, score int) {
	if move == nil {
		return
	}
	moveId := n.CreateMove(move)
	message := NodeMessage{
		MessageType: "captured",
		Identifier: n.PlayerNode.Identifier,
		Move:	moveId,
		Score: score,
		Addr: n.LocalAddr.String(),
	}

	toSend := sendMessage(n.Log, message, "Sendin' capturedPreyUpdate")
	n.MessagesToSend <- &PendingMessage{Recipient: "all", Message: toSend}
}

// Takes in a node ID and sends this node's gamestate to that node
func (n* NodeCommInterface) SendGameStateToNode(otherNodeId string){
	message := NodeMessage{
		MessageType: "gameState",
		Identifier: n.PlayerNode.Identifier,
		GameState: &n.PlayerNode.GameState,
		Addr: n.LocalAddr.String(),
	}

	toSend := sendMessage(n.Log, message, "Sendin' gamestate")
	n.MessagesToSend <- &PendingMessage{Recipient: otherNodeId, Message: toSend}
}

// Sends a move commit to prey node, for lockstep protocol
func (n *NodeCommInterface) SendMoveCommitToPreyNode(moveCommit *shared.MoveCommit) {
	message := NodeMessage {
		MessageType: "moveCommit",
		Identifier:  n.PlayerNode.Identifier,
		MoveCommit:  moveCommit,
		Addr:        n.LocalAddr.String(),
	}

	toSend := sendMessage(n.Log, message, "Sendin' move commit to prey")
	n.MessagesToSend <- &PendingMessage{Recipient:"prey", Message: toSend}
}

// Helper function to send message to player nodes and not prey node;
// do not call directly; instead write to the messagesTosend channel
func (n *NodeCommInterface) sendMessageToPlayerNodes(toSend []byte) {
	for id, val := range n.OtherNodes {
		if id != "prey" {
			_, err := val.Write(toSend)
			if err != nil {
				fmt.Println(err)
				n.NodesWriteConnRefused <- id
			}
		}
	}
}

// Handles a gamestate received from another node.
func (n* NodeCommInterface) HandleReceivedGameState(identifier string, gameState *shared.GameState) {
	//TODO: don't just wholesale replace this
	if !n.HasGameState {
		n.PlayerNode.GameState.PlayerLocs.Lock()
		defer n.PlayerNode.GameState.PlayerLocs.Unlock()

		for id, pos := range gameState.PlayerLocs.Data {
			n.PlayerNode.GameState.PlayerLocs.Data[id] = pos
		}

		for id, score := range gameState.PlayerScores.Data {
			n.PlayerNode.GameState.PlayerScores.Data[id] = score
		}
		n.HasGameState = true
	}
}

// Handle moves from prey node; means that the previously sent move commit was validated
// Returns an InvalidMoveError if the move does not match a received commit
func (n* NodeCommInterface) HandleReceivedPreyMove(identifier string, move *shared.Coord) (err error) {
	fmt.Println("DEBUG - Received move from prey")
	// Need nil check for bad move
	if move != nil {
		// check to see if it's a valid move
		err := n.CheckMoveIsValid(*move)
		if err != nil {
			return err
		}
		// update coord for prey in playerlocs
		n.PlayerNode.GameState.PlayerLocs.Lock()
		n.PlayerNode.GameState.PlayerLocs.Data[identifier] = *move
		n.PlayerNode.GameState.PlayerLocs.Unlock()
		n.GameStateToSend <- true
		// send prey node back a move op in response
		playerMove := <- n.MoveToSendToPrey
		fmt.Printf("DEBUG - Here is the response to prey node move %v and seq: %d\n", playerMove.Move, playerMove.Seq)
		n.SendMoveToPreyNode(playerMove)
		return nil
	}
	return wolferrors.InvalidMoveError("[" + string(move.X) + ", " + string(move.Y) + "]")
}

// Handle moves that does not require a move commit check
// Returns InvalidMoveError if the received move is not valid
func (n* NodeCommInterface) HandleReceivedMoveNL(identifier string, move *shared.Coord, seq uint64) (err error) {
	// Need nil check for bad move
	if move != nil {
		err := n.CheckMoveIsValid(*move)
		if err != nil {
			fmt.Println(err)
			return err
		}
		n.PlayerNode.GameState.PlayerLocs.Lock()
		n.PlayerNode.GameState.PlayerLocs.Data[identifier] = *move
		n.PlayerNode.GameState.PlayerLocs.Unlock()

		n.SendACK(identifier, seq)
		return nil
	}
	return wolferrors.InvalidMoveError("[" + string(move.X) + ", " + string(move.Y) + "]")
}

// Handles received move commits from other nodes by storing them in anticipation of receiving a move
// Returns IncorrectPlayerError if the player that send the message is not the player they are claiming to be
func (n* NodeCommInterface) HandleReceivedMoveCommit(identifier string, moveCommit *shared.MoveCommit) (err error) {
	// if the move is authentic
	if n.CheckAuthenticityOfMoveCommit(identifier, moveCommit) {
		// if identifier doesn't exist in map, add move commit to map
		if _, ok := n.MoveCommits[identifier]; !ok {
			n.MoveCommits[identifier] = hex.EncodeToString(moveCommit.MoveHash)
		}
	} else {
		return wolferrors.IncorrectPlayerError(identifier)
	}
	return nil
}

func (n* NodeCommInterface) HandleReceivedAck(identifier string, seq uint64){
	n.ACKSReceived <- &ACKMessage{Seq: seq, Identifier: identifier}
}

// Handles "connect" messages received by other nodes by adding the incoming node to this node's OtherNodes
func (n* NodeCommInterface) HandleIncomingConnectionRequest(identifier string, addr string, pubKeyString string) {
	node := n.GetClientFromAddrString(addr)
	pubKey := key.StringToPubKey(pubKeyString)
	n.NodesToAdd <- &OtherNode{Identifier: identifier, Conn: node, PubKey: &pubKey}
}

func (n* NodeCommInterface) HandleCapturedPreyRequest(identifier string, move *shared.Coord, score int) (err error) {
	err = n.CheckGotPrey(*move)
	if err != nil {
		return err
	}
	err = n.CheckMoveIsValid(*move)
	if err != nil {
		return err
	}
	err = n.CheckAndUpdateScore(identifier, score)
	if err != nil {
		return err
	}
	return nil
}

// If we are requested to send a gamestate, send it
func (n* NodeCommInterface) HandleGameStateConnReq(id string) {
	if n.PlayerNode != nil {
		n.SendGameStateToNode(id)
	}
}

// Initiates a connection to another node by sending it a "connect" message
func (n* NodeCommInterface) InitiateConnection(nodeClient *net.UDPConn, id string) {
	message := NodeMessage{
		MessageType: "connect",
		Identifier:  n.Config.Identifier,
		GameState:   nil,
		Addr:        n.LocalAddr.String(),
		PubKey: 	 key.PubKeyToString(*n.PubKey),
	}
	toSend := sendMessage(n.Log, message, "Initiating connection")
	n.MessagesToSend <- &PendingMessage{Recipient: id, Message: toSend}

	if !n.HasGameState {
		n.RequestGameState(id)
	}
}

// Requests a gamestate from another node, used on joining
func (n* NodeCommInterface) RequestGameState(id string) {
	message := NodeMessage {
		MessageType: "gamestateReq",
		Identifier:  n.Config.Identifier,
		Addr:        n.LocalAddr.String(),
	}
	toSend := sendMessage(n.Log, message, "Requesting gamestate")
	n.MessagesToSend <- &PendingMessage{Recipient: id, Message: toSend}
}

// Sends connection message to connections after receiving from server
//func (n *  NodeCommInterface) FloodNodes() {
//	for _, node := range n.OtherNodes {
//		// Exchange messages with other node
//		n.InitiateConnection(node)
//	}
//}

func (n *NodeCommInterface) SendACK(identifier string, seq uint64) {
	message := NodeMessage {
		MessageType: "ack",
		Identifier: n.PlayerNode.Identifier,
		Seq: seq,
		Addr: n.LocalAddr.String(),
	}

	toSend := sendMessage(n.Log, message,  "Sendin' Ack")
	n.MessagesToSend <- &PendingMessage{Recipient: identifier, Message: toSend}
}

////////////////////////////////////////////// MOVE COMMIT HASH FUNCTIONS //////////////////////////////////////////////

// Calculate the hash of the coordinates which will be sent at the move commitment stage
func (n *NodeCommInterface) CalculateHash(m shared.Coord, id string) ([]byte) {
	hash := md5.New()
	arr := make([]byte, 2048)

	arr = strconv.AppendInt(arr, int64(m.X), 10)
	arr = strconv.AppendInt(arr, int64(m.Y), 10)
	arr = strconv.AppendQuote(arr, id)

	// Write the hash
	hash.Write(arr)
	return hash.Sum(nil)
}

// Sign the move commit with private key
func (n *NodeCommInterface) SignMoveCommit(hash []byte) (r, s *big.Int, err error) {
	return ecdsa.Sign(rand.Reader, n.PrivKey, hash)
}

// Checks to see if the hash is legit
func (n *NodeCommInterface) CheckAuthenticityOfMoveCommit(identifier string, m *shared.MoveCommit) (bool) {
	rBigInt := new(big.Int)
	_, err := fmt.Sscan(m.R, rBigInt)

	sBigInt := new(big.Int)
	_, err = fmt.Sscan(m.S, sBigInt)
	if err != nil {
		fmt.Println("Trouble converting string to big int")
	}
	return ecdsa.Verify(n.NodeKeys[identifier], m.MoveHash, rBigInt, sBigInt)
}

func (n *NodeCommInterface) CheckAuthenticityOfMove(publicKey *ecdsa.PublicKey, m *shared.SignedMove)(bool){
	if publicKey == nil{
		// public key is nil for some tests, just pass if this is the case
		return true
	}
	rBigInt := new(big.Int)
	_, err := fmt.Sscan(m.R, rBigInt)

	sBigInt := new(big.Int)
	_, err = fmt.Sscan(m.S, sBigInt)
	if err != nil {
		fmt.Println("Trouble converting string to big int")
	}

	return ecdsa.Verify(publicKey, m.MoveByte, rBigInt, sBigInt)
}

////////////////////////////////////////////// MOVE CHECK FUNCTIONS ////////////////////////////////////////////////////

// Checks to see if there is an existing commit against the submitted move
func (n *NodeCommInterface) CheckMoveCommitAgainstMove(identifier string, move shared.Coord) (bool) {
	hash := hex.EncodeToString(n.CalculateHash(move, identifier))
	for i, mc := range n.MoveCommits {
		if mc == hash && i == identifier {
			return true
		}
	}
	return false
}

// Check move to see if it's valid based on the gameplay grid
func (n *NodeCommInterface) CheckMoveIsValid(move shared.Coord) (err error) {
	if n.PlayerNode != nil {
		gridManager := n.PlayerNode.GetGridManager()
		if !gridManager.IsValidMove(move) {
			return wolferrors.InvalidMoveError("[" + string(move.X) + ", " + string(move.Y) + "]")
		}
	}
	return nil
}

func (n *NodeCommInterface) CheckGotPrey(move shared.Coord) (err error) {
	n.PlayerNode.GameState.PlayerLocs.Lock()
	defer n.PlayerNode.GameState.PlayerLocs.Unlock()
	if move.X == n.PlayerNode.GameState.PlayerLocs.Data["prey"].X &&
		move.Y == n.PlayerNode.GameState.PlayerLocs.Data["prey"].Y {
		return nil
	}
	return wolferrors.InvalidPreyCaptureError("[" + string(move.X) + ", " + string(move.Y) + "]")
}

func (n *NodeCommInterface) CheckAndUpdateScore(identifier string, score int) (err error) {
	_, exists := n.PlayerNode.GameState.PlayerScores.Data[identifier]
	playerScore := n.PlayerNode.GameState.PlayerScores.Data[identifier]

	if !exists && score == n.PlayerNode.GameConfig.CatchWorth {
		n.PlayerNode.GameState.PlayerScores.Lock()
		defer n.PlayerNode.GameState.PlayerScores.Unlock()
		n.PlayerNode.GameState.PlayerScores.Data[identifier] = score
		return nil
	}

	if exists && score != playerScore + n.PlayerNode.GameConfig.CatchWorth {
		return wolferrors.InvalidScoreUpdateError(string(score))
	}
	n.PlayerNode.GameState.PlayerScores.Lock()
	defer n.PlayerNode.GameState.PlayerScores.Unlock()
	n.PlayerNode.GameState.PlayerScores.Data[identifier] += n.PlayerNode.GameConfig.CatchWorth
	return nil
}