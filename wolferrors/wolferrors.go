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
	return fmt.Sprintf("WolfPack: [%s]", string(e))
}

type OutOfBoundsError string

func (e OutOfBoundsError) Error() string {
	return fmt.Sprintf("WolfPack: [%s]", string(e))
}

type InvalidNonceError string

func (e InvalidNonceError) Error() string {
	return fmt.Sprintf("WolfPack: cannot create move commit hash, nonce [%d] invalid PoW", e)
}

type InvalidScoreUpdateError string

func (e InvalidScoreUpdateError) Error() string {
	return fmt.Sprintf("WolfPack: score update [%d] is incorrect", e)
}

type InvalidPreyCaptureError string

func (e InvalidPreyCaptureError) Error() string {
	return fmt.Sprintf("WolfPack: prey was not captured")
}

type IncorrectPlayerError string

func (e IncorrectPlayerError) Error() string {
	return fmt.Sprintf("WolfPack: Hash [%d] was sent by incorrect player", e)
}