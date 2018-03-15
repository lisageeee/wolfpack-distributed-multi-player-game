/*

This package specifies the application's interface to the
library (wolflib) to be used for project 2

*/

package wolflib

import (
	"crypto/ecdsa"
	"../shared"
)
// Current design is for the player app to live outside of Azure, so we would be making
// RPC calls to its compadre node to get info to render on screen
type WolfGame interface {
	// Returns the game states from other players via node on Azure.
	// Can return the following errors:
	// - DisconnectedError
	// - NoMoveCommitError
	GetGameStates(pubKey string, gameStates *[]shared.GameState) (err error)

	// Sends the coordinates of this player's sprite to node on Azure.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidMoveError
	// - OutOfBoundsError
	MoveSprite(move shared.GameState, isValid *bool) (err error)
}

// Constructor for a new Game object instance. Takes the comm node's IP:port address
// string and a public-private key pair (ecdsa private key type contains the public key).
// Returns a WolfGame instance that can be used for all future interactions with wolflib.

// Can return the following errors:
// - DisconnectedError
func startGame(commNodeAddr string, privKey ecdsa.PrivateKey) (wolfGame WolfGame,
	initSettings shared.InitialGameSettings, err error) {
	return
}