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

func TestNodeToNodeSendMove(t *testing.T) {
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

	// Test sending a move from one node to another
	testCoord := shared.Coord{7,7}
	n1.SendMoveToNodes(&testCoord)
	time.Sleep(100*time.Millisecond)

	if n2.PlayerNode.GameState.PlayerLocs[node1.Identifier] != testCoord {
		fmt.Println("Should have updated n1's location in n2's game state")
		t.Fail()
	}

	testCoord = shared.Coord{6,3}
	n2.SendMoveToNodes(&testCoord)
	time.Sleep(100*time.Millisecond)

	if n1.PlayerNode.GameState.PlayerLocs[node2.Identifier] != testCoord {
		fmt.Println("Should have updated n2's locatio in n1's game state")
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestNodeToNodeSendingNilMove(t *testing.T) {
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
	n1.SendMoveToNodes(nil)
	time.Sleep(100*time.Millisecond)

	if _, ok := n2.PlayerNode.GameState.PlayerLocs.Data[node1.Identifier]; ok {
		fmt.Println("Expected n2 to NOT contain n1's identifier in PlayerLocs map")
		t.Fail()
	}

	fmt.Printf("Node 2's PlayerLocs map %v\n", n2.PlayerNode.GameState.PlayerLocs)

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestNodeToNodeValidSelfMove(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":12830", ":12831", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":12940", ":12941", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node3 := l.CreatePlayerNode(":12950", ":12951", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node4 := l.CreatePlayerNode(":12960", ":12961", pub, priv, ":8081")

	n1 := node1.GetNodeInterface()
	n2 := node2.GetNodeInterface()
	n3 := node3.GetNodeInterface()
	n4 := node4.GetNodeInterface()

	time.Sleep(1*time.Second)

	// Test sending a move from one node to another
	testCoord := shared.Coord{7,7}
	n1.SendMoveToNodes(&testCoord)
	time.Sleep(3*time.Second)

	if n2.PlayerNode.GameState.PlayerLocs[node1.Identifier] != testCoord {
		fmt.Println("Should have updated n1's location in n2's game state")
		t.Fail()
	}

	if n3.PlayerNode.GameState.PlayerLocs[node1.Identifier] != testCoord {
		fmt.Println("Should have updated n1's location in n2's game state")
		t.Fail()
	}

	if n4.PlayerNode.GameState.PlayerLocs[node1.Identifier] != testCoord {
		fmt.Println("Should have updated n1's location in n2's game state")
		t.Fail()
	}

	time.Sleep(5*time.Second)

	if n1.PlayerNode.GameState.PlayerLocs[n1.PlayerNode.Identifier] != testCoord {
		fmt.Printf("Should have updated n1's coords to %v, instead it's %v\n", testCoord,
			n1.PlayerNode.GameState.PlayerLocs[n1.PlayerNode.Identifier])
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestPruningNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3 * time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":17830", ":17831", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":18940", ":18941", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node3 := l.CreatePlayerNode(":19950", ":19951", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node4 := l.CreatePlayerNode(":20950", ":20951", pub, priv, ":8081")

	n1 := node1.GetNodeInterface()
	n2 := node2.GetNodeInterface()
	_ = node3.GetNodeInterface()
	_ = node4.GetNodeInterface()

	n1.Strikes.StrikeCount[n2.PlayerNode.Identifier] = l.STRIKE_OUT
	time.Sleep(1 * time.Second)

	// Test sending a move from one node to another
	testCoord := shared.Coord{7, 7}
	n1.SendMoveToNodes(&testCoord)

	time.Sleep(10 * time.Second)

	fmt.Printf("Here is the Strike Count map for n1: %v\n", n1.Strikes.StrikeCount)

	if _, ok := n1.Strikes.StrikeCount[node2.Identifier]; ok {
		fmt.Println("This node should not be in the strikes map")
		t.Fail()
	}

	if _, ok := n1.OtherNodes[node2.Identifier]; ok {
		fmt.Println("This node should not be in the other nodes map")
		t.Fail()
	}
}