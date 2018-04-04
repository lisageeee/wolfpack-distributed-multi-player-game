package test

import (
	"os/exec"
	"testing"
	"context"
	"time"
	"syscall"
	key "../key-helpers"
	l "../logic/impl"
	"fmt"
)

// NOTE: eventually this test will fail, as we won't be wholesale replacing the gamestate, can remove then
func TestNodeToNodeSendGameState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":12200", ":12201", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":11900", ":11901", pub, priv, ":8081")

	n1 := node1.GetNodeInterface()
	n2 := node2.GetNodeInterface()

	time.Sleep(1*time.Second)

	// Check nodes are connected to each other
	if len(n2.OtherNodes) != len(n1.OtherNodes) {
		fmt.Println("Nodes do not have a mutual connection, fail")
		t.Fail()
	}

	// Test sending gamestate from one node to another
	n1.SendGameStateToNode(node2.Identifier)
	time.Sleep(300*time.Millisecond)

	_, ok := n2.PlayerNode.GameState.PlayerLocs.Data[n1.PlayerNode.Identifier]
	fmt.Println(n2.PlayerNode.GameState.PlayerLocs, n1.PlayerNode.GameState.PlayerLocs)
	if !ok {
		fmt.Println("Gamestate not sent from 1 to 2, fail")
		t.Fail()
	}

	if len(n2.PlayerNode.GameState.PlayerLocs.Data) != len(n1.PlayerNode.GameState.PlayerLocs.Data) + 1 {
		fmt.Println("P2 should have player 1's gamestate + its own location:", n2.PlayerNode.GameState.PlayerLocs.Data,
			n1.PlayerNode.GameState.PlayerLocs.Data)
		t.Fail()
	}

	n1.SendGameStateToNode(node2.Identifier)
	time.Sleep(100*time.Millisecond)

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}