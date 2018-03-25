package test

import (
	"testing"
	l "../logic/impl"
	p "../pixel/impl"
	"context"
	"time"
	"os/exec"
	"../shared"
	"fmt"
	"syscall"
	key "../key-helpers"
)

//NOTE command line args for playerNode:
//nodeListenerAddr = os.Args[1]
//playerListenerIpAddress = os.Args[2]
//pixelIpAddress = os.Args[3]


// Reference for killing exec.Command processes + childen:
// https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773

// This test will fail if you make a breaking change that keeps pixel.go from running
// Inspiration: the breaking change I added that prevented pixel.go from running (wrong image path)
func TestPixelNodeCanRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	fmt.Println(-serverStart.Process.Pid)
	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	_ = l.CreatePlayerNode(":12500", ":12501", ":12502", pub, priv, ":8081")

	pixelStart := exec.Command("go", "run", "pixel.go", ":12401", ":12402")
	pixelStart.Dir = "../pixel"
	pixelStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var err error

	// Start the pixel node, set err to the error returned if any
	go func() {
		_, err = pixelStart.Output()
	}()

	// Wait 5 seconds for errors to return
	time.Sleep(5 * time.Second)

	// If pixel can't start, will get err on this line
	if err != nil {
		fmt.Println("Pixel couldn't start, error:", err)
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-pixelStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()

}

// Tests that the logic node is able to send messages to the pixel node
func TestLogicNodeToPixelComm(t *testing.T) {
	// Start running server
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started

	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	n := l.CreatePlayerNode(":12300", ":12301", ":12302", pub, priv, ":8081")
	remote := n.GetPixelInterface()

	//Run pixel node
	pixel := p.CreatePixelNode(":12301", ":12302")

	// Create a gameState
	//gameState := shared.GameRenderState {
	//	PlayerLoc: shared.Coord{2,1},
	//	Prey: shared.Coord{0,1},
	//	OtherPlayers: make(map[string]shared.Coord),
	//}

	// Create a gamestate
	playerLocs := make(map[string]shared.Coord)
	playerLocs[n.Identifier] = shared.Coord{2,1}
	playerLocs["prey"] = shared.Coord{0,1}
	gameState := shared.GameState{
		PlayerLocs: playerLocs,
	}

	// Send from the remote interface to the pixel node
	remote.SendPlayerGameState(gameState, n.Identifier)

	// Read from the pixel node's channel
	pixelGameState := <- pixel.NewGameStates

	// Check it was sent correctly
	if pixelGameState.PlayerLoc.X != gameState.PlayerLocs[n.Identifier].X {
		t.Fail()
	}
	if pixelGameState.PlayerLoc.Y != gameState.PlayerLocs[n.Identifier].Y {
		t.Fail()
	}

	serverStart.Process.Kill()
}

// Tests that the pixel node can send messages to the logic node
func TestPixelNodeMove(t *testing.T) {
	// Start running server
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.Dir = "../server"
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Start()

	time.Sleep(2 * time.Second) // wait for server to get started

	// Create player node, run it and get pixel interface
	pub, priv := key.GenerateKeys()
	n := l.CreatePlayerNode(":12303", ":12304", ":12305", pub, priv, ":8081")
	go n.RunGame()
	loc := n.GameState.PlayerLocs[n.Identifier] // get the initial game render state

	// Start pixel node
	pixel := p.CreatePixelNode(":12304", ":12305")

	pixel.SendMove("up")

	// Wait a tick for the move to be sent
	time.Sleep(100*time.Millisecond)

	// Check that the player has moved up
	newState := n.GameState

	fmt.Println(newState, loc)
	if newState.PlayerLocs[n.Identifier].X != loc.X {
		t.Fail()
	}
	if newState.PlayerLocs[n.Identifier].Y - 1 != loc.Y {
		t.Fail()
	}

	// Reset to try moving down
	loc = newState.PlayerLocs[n.Identifier]
	pixel.SendMove("down")

	// Wait a tick for the move to be sent
	time.Sleep(100*time.Millisecond)

	// Check that the player has moved down
	newState = n.GameState
	fmt.Println(newState, loc)
	if newState.PlayerLocs[n.Identifier].X != loc.X {
		t.Fail()
	}
	if newState.PlayerLocs[n.Identifier].Y != loc.Y - 1 {
		t.Fail()
	}
	serverStart.Process.Kill()
}
