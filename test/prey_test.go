package test

import (
	"os/exec"
	"testing"
	"context"
	"time"
	"syscall"
	key "../key-helpers"
	l "../prey/impl"
	l2 "../logic/impl"
	"fmt"
)

func TestPreyNodeToNodeInterface(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePreyNode(":17700", ":17701", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node2 := l2.CreatePlayerNode(":17900", ":17901", pub, priv, ":8081")

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
	time.Sleep(100*time.Millisecond)

	_, ok := n2.PlayerNode.GameState.PlayerLocs.Data["prey"]
	fmt.Println(n2.PlayerNode.GameState.PlayerLocs, n1.PreyNode.GameState.PlayerLocs)
	if !ok {
		fmt.Println("Gamestate not sent from 1 to 2, fail")
		t.Fail()
	}

	if len(n2.PlayerNode.GameState.PlayerLocs.Data) != len(n1.PreyNode.GameState.PlayerLocs.Data) {
		fmt.Println("Gamestates not equal length (so not equal), fail")
		t.Fail()
	}

	n1.SendGameStateToNode(node2.Identifier)
	time.Sleep(100*time.Millisecond)

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestMovePrey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started

	pub, priv := key.GenerateKeys()
	preyNode := l.CreatePreyNode(":12210", ":12211", pub, priv, ":8081")

	// Initial coordinates @ 5,5
	loc := preyNode.MovePrey("up")
	if loc.X != 5 && loc.Y != 6 {
		t.Fail()
	}
	loc = preyNode.MovePrey("down")
	if loc.X != 5 && loc.Y != 5 {
		t.Fail()
	}
	loc = preyNode.MovePrey("right")
	if loc.X != 6 && loc.Y != 5 {
		t.Fail()
	}
	loc = preyNode.MovePrey("left")
	if loc.X != 5 && loc.Y != 5 {
		t.Fail()
	}
}