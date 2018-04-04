package test

import (
	"testing"
	l "../logic/impl"
	p "../pixel/impl"
	keys "../key-helpers"
	"time"
	"os/exec"
	"syscall"
	"context"
	"../shared"
	"fmt"
)

func TestGetNewRenderState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go", "9099")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second)
	pub, priv := keys.GenerateKeys()
	logic := l.CreatePlayerNode(":0", ":7060", pub, priv, ":9099")
	logicPoint := &logic
	go logic.RunGame(":7060")
	time.Sleep(1*time.Second)
	pixel := p.CreatePixelNode(":7060")
	go pixel.RunRemoteNodeListener()

	time.Sleep(2*time.Second)

	// Create a gamestate
	playerLocs := make(map[string]shared.Coord)
	playerLocs[logic.Identifier] = shared.Coord{18,18}
	playerLocs["cool-test-node"] = shared.Coord{15,15}
	playerLocs["prey"] = shared.Coord{12,12}

	// Assign it to the logic node
	logicPoint.GameState.PlayerLocs.Lock()
	logicPoint.GameState.PlayerLocs.Data = playerLocs
	logicPoint.GameState.PlayerLocs.Unlock()
	fmt.Println("LOgic gamesate:", &logic.GameState)

	// Send pixel the state
	interf := logicPoint.GetPixelInterface()
	pointer := &interf
	pointer.SendPlayerGameState(logicPoint.GameState)

	renderState := <-pixel.NewGameStates
	fmt.Println(renderState)

	if len(renderState.OtherPlayers) != 1 {
		fmt.Printf("Render state doesn't have expected number of other players: expected 1, got %v", len(renderState.OtherPlayers))
		t.Fail()
	}

	// Remove the other player
	_, ok := logicPoint.GameState.PlayerLocs.Data["cool-test-node"]
	if ok {
		delete(logicPoint.GameState.PlayerLocs.Data, "cool-test-node")
		fmt.Println(logicPoint.GameState.PlayerLocs.Data)
	} else {
		fmt.Printf("Render state doesn't have the expected node")
		t.Fail()
	}

	time.Sleep(1*time.Second)

	// Send pixel the state
	pointer.SendPlayerGameState(logicPoint.GameState)

	renderState = <-pixel.NewGameStates
	fmt.Println("Received gamestate", renderState.OtherPlayers)

	if len(renderState.OtherPlayers) != 0 {
		fmt.Printf("Render state doesn't have expected number of other players: expected 0, got [%v]", len(renderState.OtherPlayers))
		t.Fail()
	}


	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}