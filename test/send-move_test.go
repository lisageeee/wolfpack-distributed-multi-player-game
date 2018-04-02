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

	if n2.PlayerNode.GameState.PlayerLocs.Data[node1.Identifier] != testCoord {
		fmt.Println("Should have updated n1's location in n2's game state")
		t.Fail()
	}

	testCoord = shared.Coord{6,3}
	n2.SendMoveToNodes(&testCoord)
	time.Sleep(100*time.Millisecond)

	if n1.PlayerNode.GameState.PlayerLocs.Data[node2.Identifier] != testCoord {
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

	time.Sleep(1 * time.Second)

	fmt.Printf("Length of other nodes: %d\n", len(n1.OtherNodes))

	// Test sending a move from one node to another
	testCoord := shared.Coord{3, 7}
	n2.SendMoveToNodes(&testCoord)
	time.Sleep(3*time.Second)

	n1.PlayerNode.GameState.PlayerLocs.RLock()
	if _, ok := n1.PlayerNode.GameState.PlayerLocs.Data[n2.PlayerNode.Identifier]; !ok {
		fmt.Println("n2's loc has not been captured")
		t.Fail()
	}
	n1.PlayerNode.GameState.PlayerLocs.RUnlock()

	// Test sending a move from one node to another
	testCoord = shared.Coord{7, 7}
	n1.SendMoveToNodes(&testCoord)
	n1.SendMoveToNodes(&testCoord)
	n1.SendMoveToNodes(&testCoord)

	n2.HeartAttack <- true

	time.Sleep(10*time.Second)

	n1.Strikes.RLock()
	fmt.Printf("Here is the Strike Count map for n1: %v\n", n1.Strikes.StrikeCount)
	for _, v := range n1.Strikes.StrikeCount {
		if v > l.STRIKE_OUT {
			fmt.Println("There's a node with a strike out greater than 3")
			t.Fail()
		}
	}
	n1.Strikes.RUnlock()

	n1.PlayerNode.GameState.PlayerLocs.RLock()
	if n1.PlayerNode.GameState.PlayerLocs.Data[n1.PlayerNode.Identifier] != testCoord {
		fmt.Println("n1's node has not been updated")
		t.Fail()
	}
	if _, ok := n1.PlayerNode.GameState.PlayerLocs.Data[n2.PlayerNode.Identifier]; !ok {
		fmt.Println("n2's loc has not been deleted")
		t.Fail()
	}
	n1.PlayerNode.GameState.PlayerLocs.RUnlock()

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}