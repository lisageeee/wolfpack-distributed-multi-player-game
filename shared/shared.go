package shared

import (
	_ "crypto/ecdsa"
	"sync"
	"net"
)

// Coordinates of an element in game
type Coord struct {
	X int
	Y int
}

type GameConfig struct {
	InitState			InitialState
	Identifier 			string
	GlobalServerHB		uint32
	// Number of times we ping another player before we drop them
	Ping				uint32
}

// Initial game settings sent out by global server to start the game
type InitialGameSettings struct {
	WindowsX			float64
	WindowsY			float64
	WallCoordinates		[]Coord
	ScoreboardWidth		float64
}

type InitialState struct {
	Settings 	InitialGameSettings
	CatchWorth	int
}
// Game state sent by other player, or from this player
type PlayerState struct {
	PlayerId			uint32
	PlayerLoc			Coord
	Timestamp			uint64
	LastUpdated			uint64
	HighestScore 		uint32
}

// Game state to communiciate between nodes
type GameState struct {
	PlayerLocs 		PlayerLockMap
	PlayerScores	ScoresLockMap
}

type PlayerLockMap struct {
	sync.RWMutex
	Data map[string]Coord
}

type ScoresLockMap struct {
	sync.RWMutex
	Data map[string]int
}

// Game state sent from logic node to pixel for rendering
type GameRenderState struct {
	PlayerLoc Coord
	Prey Coord
	OtherPlayers map[string]Coord
	Scores map[string]int
}

// Move commitment sent by player, must be ACK'ed by all other players in game
// before this player can receive all other players' game states
type MoveCommit struct {
	Seq					uint64
	MoveHash			[]byte
	R					string
	S					string
}

type MoveOp struct {
	PlayerLoc     		Coord
	PlayerId			string
}

type SignedMove struct{
	MoveByte		    []byte
	R					string
	S					string
}

// A struct to communicate between the server and other nodes and also between nodes the identification details of
// a player node; includes identifier, public key, and address
type NodeRegistrationInfo struct {
	Id string
	Addr net.Addr
	PubKey string
}
