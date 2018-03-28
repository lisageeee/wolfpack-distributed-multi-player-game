package test

import "testing"
import "../shared"
import (
	"../wolfnode"
	"../key-helpers"
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

