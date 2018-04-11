package test

import (
	"testing"
	"time"
	"os/exec"
	"syscall"
	"fmt"
	key "../key-helpers"
	l "../logic/impl"
	"../shared"
	"context"
)

func TestValidPreyCapture (t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":123456", ":16435", pub, priv, ":8081")

	node1.GameState.PlayerLocs.Data["prey"] = shared.Coord{6, 6}
	node1.GameState.PlayerLocs.Data[node1.Identifier] = shared.Coord{6,6}

	err := node1.GetNodeInterface().CheckGotPrey(node1.GameState.PlayerLocs.Data[node1.Identifier])
	if err != nil {
		t.Fail()
	}

	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestInvalidPreyCapture (t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":19191", ":11919", pub, priv, ":8081")

	node1.GameState.PlayerLocs.Data["prey"] = shared.Coord{8, 6}
	node1.GameState.PlayerLocs.Data[node1.Identifier] = shared.Coord{6,6}

	err := node1.GetNodeInterface().CheckGotPrey(node1.GameState.PlayerLocs.Data[node1.Identifier])
	if err == nil {
		t.Fail()
	}

	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestNodeToNodeSendScore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":12820", ":12821", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":12920", ":12921", pub, priv, ":8081")

	n1 := node1.GetNodeInterface()
	n2 := node2.GetNodeInterface()

	time.Sleep(1*time.Second)

	// Check nodes are connected to each other
	if len(n2.OtherNodes) != len(n1.OtherNodes) {
		fmt.Println("Nodes do not have a mutual connection, fail")
		t.Fail()
	}

	// Test sending a score from one node to another
	testCoord := shared.Coord{5,5}
	node2.GameState.PlayerLocs.Data["prey"] = shared.Coord{5, 5}

	err := n2.HandleCapturedPreyRequest(node1.Identifier, &testCoord, 1, uint64(9))
	if err != nil {
		fmt.Println("Error in sending a valid prey & score")
		fmt.Println(err)
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestNodeToNodeSendingInvalidScore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":14730", ":14731", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":11730", ":11731", pub, priv, ":8081")

	n1 := node1.GetNodeInterface()
	n2 := node2.GetNodeInterface()

	time.Sleep(1*time.Second)

	// Check nodes are connected to each other
	if len(n2.OtherNodes) != len(n1.OtherNodes) {
		fmt.Println("Nodes do not have a mutual connection, fail")
		t.Fail()
	}

	// Test sending a move from one node to another
	testCoord := shared.Coord{7,1}
	n1.SendPreyCaptureToNodes(&testCoord, 1)
	time.Sleep(200*time.Millisecond)

	err := n2.HandleCapturedPreyRequest(node1.Identifier, &testCoord, 4, uint64(6))
	if err == nil {
		fmt.Println("Error in sending an invalid prey & score 1")
		fmt.Println(err)
		t.Fail()
	}

	testCoord = shared.Coord{5,5}
	n1.SendPreyCaptureToNodes(&testCoord, 3)
	time.Sleep(300*time.Millisecond)

	err = n2.HandleCapturedPreyRequest(node1.Identifier, &testCoord, 3, uint64(2))
	if err == nil {
		fmt.Println("Error in sending an invalid prey & score 2")
		fmt.Println(err)
		t.Fail()
	}
	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}