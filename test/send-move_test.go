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

func TestNodeToNodeSendMoveInvalidWithoutMoveCommit(t *testing.T) {
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

	if n2.PlayerNode.GameState.PlayerLocs[node1.Identifier] == testCoord {
		fmt.Println("Should not have passed the move commit check")
		t.Fail()
	}

	testCoord = shared.Coord{6,3}
	n2.SendMoveToNodes(&testCoord)
	time.Sleep(100*time.Millisecond)

	if n1.PlayerNode.GameState.PlayerLocs[node2.Identifier] == testCoord {
		fmt.Println("Should not have passed the move commit check")
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestNodeToNodeSendMoveValidWithoutMoveCommit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":14720", ":14721", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":11720", ":11721", pub, priv, ":8081")

	n1 := node1.GetNodeInterface()
	n2 := node2.GetNodeInterface()

	time.Sleep(1*time.Second)

	// Check nodes are connected to each other
	if len(n2.OtherNodes) != len(n1.OtherNodes) {
		fmt.Println("Nodes do not have a mutual connection, fail")
		t.Fail()
	}

	testCoord := shared.Coord{7,7}

	// Send move commit first
	hash1 := l.CalculateHash(testCoord, node1.Identifier)
	r, s, err := n1.SignMoveCommit(hash1)
	if err != nil {
		fmt.Println("Couldn't sign move commit")
	}
	_, pubKey := key.Encode(n1.PrivKey, n1.PubKey)
	mc := shared.MoveCommit {
		MoveHash: hash1,
		PubKey: pubKey,
		R: r.String(),
		S: s.String(),
	}

	n1.SendMoveCommitToNodes(&mc)
	time.Sleep(100*time.Millisecond)

	// Test sending a move from one node to another
	n1.SendMoveToNodes(&testCoord)
	time.Sleep(100*time.Millisecond)

	if n2.PlayerNode.GameState.PlayerLocs[node1.Identifier] != testCoord {
		fmt.Println("Expected n2 to contain n1's updated move to [7, 7]")
		t.Fail()
	}

	fmt.Printf("Node 2's PlayerLocs map %v\n", n2.PlayerNode.GameState.PlayerLocs)

	if _, ok := n2.MoveCommits[node1.Identifier]; ok {
		fmt.Println("Expected no more move commit from n1 in n2.MoveCommits")
		t.Fail()
	}

	fmt.Printf("Node 2's moveCommits map %v\n", n2.MoveCommits)

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

// TODO: Why is this not deleting the move commit? Also is that the behaviour we want?
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

	testCoord := shared.Coord{7,7}

	// Send move commit first
	hash1 := l.CalculateHash(testCoord, node1.Identifier)
	r, s, err := n1.SignMoveCommit(hash1)
	if err != nil {
		fmt.Println("Couldn't sign move commit")
	}
	_, pubKey := key.Encode(n1.PrivKey, n1.PubKey)
	mc := shared.MoveCommit {
		MoveHash: hash1,
		PubKey: pubKey,
		R: r.String(),
		S: s.String(),
	}

	n1.SendMoveCommitToNodes(&mc)
	time.Sleep(100*time.Millisecond)

	// Test sending a move from one node to another
	n1.SendMoveToNodes(nil)
	time.Sleep(100*time.Millisecond)

	if _, ok := n2.PlayerNode.GameState.PlayerLocs[node1.Identifier]; ok {
		fmt.Println("Expected n2 to NOT contain n1's identifier in PlayerLocs map")
		t.Fail()
	}

	fmt.Printf("Node 2's PlayerLocs map %v\n", n2.PlayerNode.GameState.PlayerLocs)

	if _, ok := n2.MoveCommits[node1.Identifier]; ok {
		fmt.Println("Expected no more move commit from n1 in n2.MoveCommits")
		fmt.Printf("Node 2's moveCommits map %v\n", n2.MoveCommits)
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}