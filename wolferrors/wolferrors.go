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
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> interface for wolfnode and added an error def

type InvalidNonceError string

func (e InvalidNonceError) Error() string {
<<<<<<< HEAD
	return fmt.Sprintf("WolfPack: Cannot create move commit hash, nonce [%d] invalid PoW", e)
<<<<<<< HEAD
}
=======
>>>>>>> app interface with wolflib and errors
=======
}
>>>>>>> interface for wolfnode and added an error def
=======
	return fmt.Sprintf("WolfPack: cannot create move commit hash, nonce [%d] invalid PoW", e)
}

type InvalidScoreUpdateError string

func (e InvalidScoreUpdateError) Error() string {
	return fmt.Sprintf("WolfPack: score update [%d] is incorrect", e)
}
>>>>>>> additional stubs to interfaces
