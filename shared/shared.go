package shared

import (
	"crypto/ecdsa"
	"math/big"
)

// Coordinates of an element in game
type Coord struct {
	X int
	Y int
}

type GameConfig struct {
	InitState			InitialState
	Identifier 			int
	GlobalServerHB		uint32
	// Number of times we ping another player before we drop them
	Ping				uint32
}

// Initial game settings sent out by global server to start the game
type InitialGameSettings struct {
	WindowsX			float64
	WindowsY			float64
	WallCoordinates		[]Coord
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

// Game state to communciate between nodes
type GameState struct {
	PlayerLocs map[string]Coord
	// scores TODO
}

// Game state sent from logic node to pixel for rendering
type GameRenderState struct {
	PlayerLoc Coord
	Prey Coord
	OtherPlayers map[string]Coord
}

// Move commitment sent by player, must be ACK'ed by all other players in game
// before this player can receive all other players' game states

type MoveOp struct {
	PlayerState     	PlayerState
	PubKey         		*ecdsa.PublicKey
	Signature			Sig
}

// Signature generated by ecdsa for the hashing of the move commit
type Sig struct {
	R 					*big.Int
	S 					*big.Int
}

