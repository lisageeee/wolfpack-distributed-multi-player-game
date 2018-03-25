package test

import "testing"
import "../shared"
import (
	"../wolfnode"
	"../key-helpers"
	"math/big"
	"fmt"
)

func wnSetup() (wolfnode.WolfNodeImpl) {
	wn := wolfnode.WolfNodeImpl{}
	info := wolfnode.PlayerInfo{}
	settings := shared.InitialGameSettings{3000, 3000, []shared.Coord{{1, 1}, {10, 90}, {23, 99}}}
	pub, priv := key_helpers.GenerateKeys()
	info.InitGameSettings = settings
	info.PubKey = pub
	info.PrivKey = priv
	wn.Info = info
	return wn
}

func TestCapturedPreyValid(t *testing.T) {
	wn := wnSetup()

	_, publicKeyString := key_helpers.Encode(wn.Info.PrivKey, wn.Info.PubKey)
	wn.Info.CurrGameState = make(map[string]shared.PlayerState)
	wn.Info.CurrGameState[publicKeyString] = shared.PlayerState{
		PlayerLoc: shared.Coord{5, 5},
	}

	response := wn.CheckCapturedPrey()
	if response != nil {
		t.Fail()
	}
}

func TestCapturedPreyInvalid(t *testing.T) {
	wn := wnSetup()

	_, publicKeyString := key_helpers.Encode(wn.Info.PrivKey, wn.Info.PubKey)
	wn.Info.CurrGameState = make(map[string]shared.PlayerState)
	wn.Info.CurrGameState[publicKeyString] = shared.PlayerState{
		PlayerLoc: shared.Coord{6, 5},
	}

	response := wn.CheckCapturedPrey()
	if response == nil {
		t.Fail()
	}
}

func TestCheckScoreValid(t *testing.T) {
	wn := wnSetup()

	_, publicKeyString := key_helpers.Encode(wn.Info.PrivKey, wn.Info.PubKey)
	wn.Info.CurrGameState = make(map[string]shared.PlayerState)
	wn.Info.CurrGameState[publicKeyString] = shared.PlayerState{
		PlayerLoc: shared.Coord{5, 5},
	}

	response := wn.CheckScore(1)
	if response != nil {
		t.Fail()
	}
}

func TestCheckScoreInvalid(t *testing.T) {
	wn := wnSetup()

	_, publicKeyString := key_helpers.Encode(wn.Info.PrivKey, wn.Info.PubKey)
	wn.Info.CurrGameState = make(map[string]shared.PlayerState)
	wn.Info.CurrGameState[publicKeyString] = shared.PlayerState{
		PlayerLoc: shared.Coord{6, 5},
	}

	response := wn.CheckScore(1)
	if response == nil {
		t.Fail()
	}

	response = wn.CheckScore(2)
	if response == nil {
		t.Fail()
	}
}

func TestMoveValid(t *testing.T) {
	wn := wnSetup()
	coords := shared.Coord{25, 25}
	response := wn.CheckMove(coords)
	if response != nil {
		t.Fail()
	}
}

func TestMoveInvalid(t *testing.T) {
	wn := wnSetup()
	coords := shared.Coord{3001, 25}
	response := wn.CheckMove(coords)
	if response == nil {
		t.Fail()
	}
	// wall
	coords = shared.Coord{10, 90}
	response = wn.CheckMove(coords)
	if response == nil {
		t.Fail()
	}
}

func TestMoveCommitValid(t *testing.T) {
	// I am stupid and don't know how to write a test for this xoxo
}

func TestMoveCommitInvalid(t *testing.T) {
	wn := wnSetup()

	gs :=  shared.PlayerState{
		PlayerLoc: shared.Coord{5, 5},
	}
	publicKey, _ := key_helpers.GenerateKeys()
	op := shared.MoveOp{
		PlayerState:      gs,
		PubKey:         publicKey,
		Signature:      shared.Sig{R: big.NewInt(184), S: big.NewInt(3)},
	}

	response := wn.CheckMoveCommit("AHASH", op)
	fmt.Println(response)
	if response == nil {
		t.Fail()
	}
}


