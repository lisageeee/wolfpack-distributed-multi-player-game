/*

This package specifies the application's interface to the
library (wolflib) to be used for project 2

*/

package wolflib

import "crypto/ecdsa"

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
	PlayerId		uint32
	PlayerLoc		Coordinates
	Timestamp		uint64
	LastUpdated		uint64
	HighestScore 	uint32
}

type WolfGame interface {
	// Returns the game states from other players.
	// Can return the following errors:
	// - DisconnectedError
	// - NoMoveCommitError
	GetGameStates() (gameStates []GameState, err error)

	// Sends commit move hash to other players.
	// Can return the following error:
	// - DisconnectedError
	// - InvalidMoveHashError
	CommitMove(moveHash string) (err error)

	// Sends the coordinates of this player's sprite to other players.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidMoveError
	// - OutOfBoundsError
	MoveSprite(coordinates Coordinates) (err error)
}

// Constructor for a new Game object instance. Takes the comm node's IP:port address
// string and a public-private key pair (ecdsa private key type contains the public key). Returns a WolfGame instance
// that can be used for all future interactions with wolflib.

// Can return the following errors:
// - DisconnectedError
func startGame(commNodeAddr string, privKey ecdsa.PrivateKey) (wolfGame WolfGame, initGameState InitialGameState, err error) {
	// TODO
	return
}