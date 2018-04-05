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
	go node1.RunGame(":17701")

	pub, priv = key.GenerateKeys()
	node2 := l2.CreatePlayerNode(":17900", ":17901", pub, priv, ":8081")
	go node2.RunGame(":17901")

	prey := node1.GetNodeInterface()
	n2 := node2.GetNodeInterface()

	if prey.PreyNode.GameState.PlayerLocs.Data["prey"] != n2.PlayerNode.GameState.PlayerLocs.Data["prey"] {
		fmt.Println("Initial prey coords does not match up in n2")
		fmt.Printf("Prey node GameState PlayerLocs: %v, Node 2 Gamestate PlayerLocs: %v\n",
			prey.PreyNode.GameState.PlayerLocs, n2.PlayerNode.GameState.PlayerLocs)
		t.Fail()
	}
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

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestPreyId(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started

	pub, priv := key.GenerateKeys()
	preyNode := l.CreatePreyNode(":13510", ":13511", pub, priv, ":8081")

	time.Sleep(3*time.Second) // wait for server to get started

	if preyNode.Identifier != "prey" {
		fmt.Println("Not registered as prey", preyNode.Identifier)
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}