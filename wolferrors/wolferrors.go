package wolferrors

import "fmt"

type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("WolfPack: cannot connect to [%s]", string(e))
}

type NoMoveCommitError string

func (e NoMoveCommitError) Error() string {
	return fmt.Sprintf("WolfPack: cannot get game states since there is no move commit for this move [%s]", string(e))
}

type InvalidMoveHashError string

func (e InvalidMoveHashError) Error() string {
	return fmt.Sprintf("WolfPack: invalid move hash [%s]", string(e))
}

type InvalidMoveError string

func (e InvalidMoveError) Error() string {
	return fmt.Sprintf("WolfPack: invalid move [%s]", string(e))
}

type OutOfBoundsError string

func (e OutOfBoundsError) Error() string {
	return fmt.Sprintf("WolfPack: move is out of bounds [%s]", string(e))
}

type InvalidNonceError string

func (e InvalidNonceError) Error() string {
	return fmt.Sprintf("WolfPack: cannot create move commit hash, nonce [%d] invalid PoW", e)
}

type InvalidScoreUpdateError string

func (e InvalidScoreUpdateError) Error() string {
	return fmt.Sprintf("WolfPack: score update [%d] is incorrect", e)
}

type KeyAlreadyRegisteredError string

func (e KeyAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("WolfPack: player already registered [%s]", string(e))
}

type AddressAlreadyRegisteredError string

func (e AddressAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("WolfPack: player already registered [%s]", string(e))
}

type UnknownKeyError string

func (e UnknownKeyError) Error() string {
	return fmt.Sprintf("WolfPack: unknown key [%s]", string(e))
}
