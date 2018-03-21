package test

import (
	"testing"
	l "../logic/impl"
	p "../pixel/impl"
	"context"
	"time"
	"os/exec"
	"../shared"
)

//NOTE:
//nodeListenerAddr = os.Args[1]
//playerListenerIpAddress = os.Args[2]
//pixelIpAddress = os.Args[3]

// Tests that the logic node is able to send messages to the pixel node
func TestLogicNodeToPixelComm(t *testing.T) {
	// Start running server
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(2*time.Second) // wait for server to get started

	// Create player node and get pixel interface
	n := l.CreatePlayerNode(":12300", ":12301", ":12302")
	remote := n.GetPixelInterface()
	// l.CreatePlayerNode(":12303", ":12304", ":12305")

	//Run pixel node
	pixel := p.CreatePixelNode(":12301", ":12302")

	// Create a gameState
	gameState := shared.GameRenderState {
		PlayerLoc: shared.Coord{2,1},
		Prey: shared.Coord{0,1},
		OtherPlayers: make(map[string]shared.Coord),
	}

	// Send from the remote interface to the pixel node
	remote.SendPlayerGameState(gameState)

	// Read from the pixel node's channel
	pixelGameState := <- pixel.NewGameStates

	// Check it was sent correctly
	if pixelGameState.PlayerLoc.X != gameState.PlayerLoc.X {
		t.Fail()
	}
	if pixelGameState.PlayerLoc.Y != gameState.PlayerLoc.Y {
		t.Fail()
	}
}

// Tests that the pixel node can send messages to the logic node
func TestPixelNodeMove(t *testing.T) {
	// Start running server
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(2 * time.Second) // wait for server to get started

	// Create player node, run it and get pixel interface
	n := l.CreatePlayerNode(":12303", ":12304", ":12305")
	go n.RunGame()
	state := n.GameRenderState // get the initial game render state

	// Start pixel node
	pixel := p.CreatePixelNode(":12304", ":12305")

	pixel.SendMove("up")

	// Wait a tick for the move to be sent
	time.Sleep(100*time.Millisecond)

	// Check that the player has moved up
	newState := n.GameRenderState
	if newState.PlayerLoc.X != state.PlayerLoc.X {
		t.Fail()
	}
	if newState.PlayerLoc.Y - 1 != state.PlayerLoc.Y {
		t.Fail()
	}

	// Reset to try moving down
	state = newState
	pixel.SendMove("down")

	// Wait a tick for the move to be sent
	time.Sleep(100*time.Millisecond)

	// Check that the player has moved down
	newState = n.GameRenderState
	if newState.PlayerLoc.X != state.PlayerLoc.X {
		t.Fail()
	}
	if newState.PlayerLoc.Y != state.PlayerLoc.Y - 1 {
		t.Fail()
	}
}
