package solitary
import (
	"testing"
	l "../logic/impl"
	p "../pixel/impl"
	"context"
	"time"
	"os/exec"
	"fmt"
	"syscall"
	key "../key-helpers"
)

// Tests that the pixel node can send messages to the logic node
// NOTE: this test sometimes fails unless you run it alone (wooo back to that again)
func TestPixelNodeMove(t *testing.T) {
	// Start running server
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.Dir = "../server"
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Start()

	time.Sleep(3 * time.Second) // wait for server to get started

	// Create player node, run it and get pixel interface
	pub, priv := key.GenerateKeys()
	n := l.CreatePlayerNode(":12303", ":12304", pub, priv, ":8081")
	go n.RunGame(":12304")
	loc := n.GameState.PlayerLocs.Data[n.Identifier] // get the initial game render state

	time.Sleep(1*time.Second) // wait playernode to start

	// Start pixel node
	pixel := p.CreatePixelNode(":12304")
	go pixel.RunRemoteNodeListener()

	pixel.SendMove("up")

	// Wait a tick for the move to be sent
	time.Sleep(300*time.Millisecond)

	// Check that the player has moved up
	newState := n.GameState

	fmt.Println(newState, loc)
	if newState.PlayerLocs.Data[n.Identifier].X != loc.X {
		fmt.Println("Player x changed when it shouldn't have")
		t.Fail()
	}
	if newState.PlayerLocs.Data[n.Identifier].Y - 1 != loc.Y {
		fmt.Println("Player Y didn't change when it should have")
		t.Fail()
	}

	// Reset to try moving down
	loc = newState.PlayerLocs.Data[n.Identifier]
	pixel.SendMove("down")

	// Wait a tick for the move to be sent
	time.Sleep(300*time.Millisecond)

	// Check that the player has moved down
	newState = n.GameState
	fmt.Println(newState, loc)
	if newState.PlayerLocs.Data[n.Identifier].X != loc.X {
		fmt.Println("Player x changed when it shouldn't have")
		t.Fail()
	}
	if newState.PlayerLocs.Data[n.Identifier].Y != loc.Y - 1 {
		fmt.Println("Player Y didn't change when it should have")
		t.Fail()
	}
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
}