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
	wn.Info.CurrGameState = make(map[string]shared.GameState)
	wn.Info.CurrGameState[publicKeyString] = shared.GameState{
		0,
		shared.Coord{5, 5},
		0,
		0,
		0,
	}

	response := wn.CheckCapturedPrey()
	if response != nil {
		t.Fail()
	}
}

func TestCapturedPreyInvalid(t *testing.T) {
	wn := wnSetup()

	_, publicKeyString := key_helpers.Encode(wn.Info.PrivKey, wn.Info.PubKey)
	wn.Info.CurrGameState = make(map[string]shared.GameState)
	wn.Info.CurrGameState[publicKeyString] = shared.GameState{
		0,
		shared.Coord{6, 5},
		0,
		0,
		0,
	}

	response := wn.CheckCapturedPrey()
	if response == nil {
		t.Fail()
	}
}

func TestCheckScoreValid(t *testing.T) {

}

func TestCheckScoreInvalid(t *testing.T) {

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

}

func TestMoveCommitInvalid(t *testing.T) {
	wn := wnSetup()

	gs :=  shared.GameState{
		0,
		shared.Coord{5, 5},
		0,
		0,
		0,
	}
	publicKey, _ := key_helpers.GenerateKeys()
	commit := shared.MoveCommit{
		gs,
		"AHASH",
		publicKey,
		shared.Sig{R: big.NewInt(184), S: big.NewInt(3)},
	}

	response := wn.CheckMoveCommit(commit)
	fmt.Println(response)
	if response == nil {
		t.Fail()
	}
}


