package wolfnode

import (
	"net/rpc"
	"crypto/ecdsa"
	"net"
)

// Represents coordinates for elements in game
type Coordinates struct {
	X float64
	Y float64
}

// Initial game state sent out by global server to start the game
type InitialGameSettings struct {
	// TODO: Work with Lisa's implementation of it from GlobalServer
	// TODO: Should we also sent the images too for the sprites / wall?

	// Size of game screen
	WindowsX			float64
	WindowsY			float64

	// Size of player sprites
	SpriteMin			float64
	SpriteMax			float64
	SpriteStep			float64
	SpriteCoordinates 	Coordinates

	// Walls
	WallCoordinates		[]Coordinates

	// Reward for catching prey
	Points				uint32

	// Hearbeat settings
	GlobalServerHB		uint32

	// Number of times we ping another player before we drop them
	Ping				uint32
}

// Game state sent by other player, or from this player
type GameState struct {
	PlayerId			uint32
	PlayerLoc			Coordinates
	Timestamp			uint64
	LastUpdated			uint64
	HighestScore 		uint32
}

// A player information object
type playerInfo struct {
	ServerAddr       string
	ServerConn       *rpc.Client
	PubKey           ecdsa.PublicKey
	PrivKey          ecdsa.PrivateKey
	PlayerIP         net.Addr
	InitGameSettings InitialGameSettings
	CurrGameState    GameState
	// do we need this if we're using keys?
	PlayerId			uint32
	OtherPlayersConn	map[string]*rpc.Client
	// since we're using UDP to send msgs to/from player nodes, this is to track how many times we are unable to
	// reach another node. If it crosses the threshold, some number set by server?, then we delete from OtherPlayersConn
	OtherPlayersTracker	map[string]int
	MoveCommits			map[string]string
}

// Player connection details
type playerConn struct {
	PubKey				ecdsa.PublicKey
	playerIP			*rpc.Client
}

type WolfNode interface {
	// Register with server with a one-way node to server RPC connection.
	// Gets InitGameSettings, and PlayerId (this to be tbd)
	// Sends Pubkey and PlayerIP
	// Can return the following errors:
	// - DisconnectedError
	// - TODO: AlreadyRegisteredError?
	RegisterServer(serverAddr string) (err error)

	// Sets up a hearbeat protocol with the global server to let it know that this player is alive.
	// Can return the following errors:
	// - DisconnectedError
	SendHearbeatsGlobalServer()

	// Returns the other players' connection information from global server.
	// Updates this node's OtherPlayersConn attribute (add to, or delete from).
	// Can return the following errors:
	// - DisconnectedError
	GetNodes() (otherPlayers []playerConn, err error)

	// Sets up a heartbeat protocol with the other player nodes to let them know that this player is alive.
	// Can return the following errors:
	// - DisconnectedError
	SendHearbeatsOtherPlayers() (err error)

	// Updates this node's OtherPlayersConn attribute (delete from) iff we do not receive a "I'm alive" message Ping times.
	// Can return the following errors:
	// - DisconnectedError
	TrackOtherPlayersNodes() (err error)

	// TODO

	/////// MOVES ////////
	// NODE SERVICE: Send a move commit to other players
	SendMoveCommitment()

	// NODE SERVICE: Send moves to other players
	SendMove()

	// NODE SERVICE: Send updated score after capturing prey
	SendUpdatedScore()

	/////// CHECKS ///////
	// OTHER PLAYER: Receive move commit hash from another player. Check authenticity of move commit
	CheckMoveCommitHash()

	// OTHER PLAYER: Receive move from another player. Check authenticity of move commit, and see if they've previously
	// submitted a valid move commit hash
	CheckValidMove()

	// LOCAL: Check app's move to see if it's valid based on this node's game state
	CheckMove()

	// LOCAL: Check app's move to see if they actually got the prey
	CheckCapturedPrey()

	// LOCAL: Check app's update oF high score is valid
	CheckScore()
}