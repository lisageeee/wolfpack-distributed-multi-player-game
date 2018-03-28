package test

import (
	"testing"
	"fmt"
	"context"
	"time"
	"os/exec"
	"syscall"
	key "../key-helpers"
	l "../logic/impl"
	"../shared"
	"encoding/hex"
)

func TestHashAndSigning (t *testing.T) {
	pub, priv := key.GenerateKeys()
	pn := l.PlayerNode{Identifier: "test1",
	}
	n := l.NodeCommInterface {
		PlayerNode: &pn,
		PubKey: pub,
		PrivKey: priv,
	}

	hashStr := l.CalculateHash(shared.Coord{8,9}, n.PlayerNode.Identifier)
	r, s, err := n.SignMoveCommit(hashStr)
	if err != nil {
		fmt.Println("Something went wrong with signing move commit")
		t.Fail()
	}

	_, pubStr := key.Encode(priv, pub)

	mc := shared.MoveCommit{
		MoveHash: hashStr,
		PubKey: pubStr,
		R: r.String(),
		S: s.String(),
	}
	if !n.CheckAuthenticityOfMoveCommit(&mc) {
		fmt.Println("Verifying hash == false")
		t.Fail()
	}
}

func TestCheckMoveCommitAgainstMove (t *testing.T) {
	pub, priv := key.GenerateKeys()
	pn := l.PlayerNode{Identifier: "test2",
	}
	n := l.NodeCommInterface {
		PlayerNode: &pn,
		PubKey: pub,
		PrivKey: priv,
	}
	testCoords := shared.Coord{8,9}
	hashStr := l.CalculateHash(testCoords, n.PlayerNode.Identifier)
	n.MoveCommits = make(map[string]string)
	n.MoveCommits["test2"] = hex.EncodeToString(hashStr)

	if !n.CheckMoveCommitAgainstMove("test2", testCoords) {
		fmt.Println("There is no move associated with a move commit in n.MoveCommits map")
		t.Fail()
	}
}

func TestCheckMoveCommitAgainstMoveInvalid (t *testing.T) {
	pub, priv := key.GenerateKeys()
	pn := l.PlayerNode{Identifier: "test2",
	}
	n := l.NodeCommInterface {
		PlayerNode: &pn,
		PubKey: pub,
		PrivKey: priv,
	}
	testCoords := shared.Coord{8,9}
	hashStr := l.CalculateHash(testCoords, n.PlayerNode.Identifier)
	n.MoveCommits = make(map[string]string)
	n.MoveCommits["test2"] = hex.EncodeToString(hashStr)

	if n.CheckMoveCommitAgainstMove("SoMeOtHErId", testCoords) {
		fmt.Println("There should not be a matching hash in n.MoveCommits map")
		t.Fail()
	}
}

func TestNodeToNodeSendMoveCommit (t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":15800", ":15801", pub, priv, ":8081")

	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":15900", ":15901", pub, priv, ":8081")

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
	hash := l.CalculateHash(testCoord, n1.PlayerNode.Identifier)
	r, s, err := n1.SignMoveCommit(hash)
	if err != nil {
		fmt.Println("Error signing hash, fail")
		t.Fail()
	}
	_, pubKey := key.Encode(n1.PrivKey, n1.PubKey)
	moveCommit := shared.MoveCommit {
		MoveHash: hash,
		PubKey: pubKey,
		R: r.String(),
		S: s.String(),
	}
	n1.SendMoveCommitToNodes(&moveCommit)
	time.Sleep(1*time.Second)

	if n2.MoveCommits[node1.Identifier] != hex.EncodeToString(moveCommit.MoveHash) {
		fmt.Println("Failed to send hash from Node 1 to node 2")
		fmt.Printf("Node 2 [%v]\n", n2.MoveCommits[node1.Identifier])
		fmt.Printf("Hash sent [%v]\n", hex.EncodeToString(moveCommit.MoveHash))
		t.Fail()
	}

	fmt.Printf("Node 2 [%v]\n", n2.MoveCommits[node1.Identifier])
	fmt.Printf("Hash sent [%v]\n", hex.EncodeToString(moveCommit.MoveHash))

	testCoord = shared.Coord{6,3}
	hash = l.CalculateHash(testCoord, n2.PlayerNode.Identifier)
	r, s, err = n2.SignMoveCommit(hash)
	if err != nil {
		fmt.Println("Error signing hash, fail")
		t.Fail()
	}
	_, pubKey = key.Encode(n2.PrivKey, n2.PubKey)
	moveCommit2 := shared.MoveCommit {
		MoveHash: hash,
		PubKey: pubKey,
		R: r.String(),
		S: s.String(),
	}
	n2.SendMoveCommitToNodes(&moveCommit2)
	time.Sleep(1*time.Second)

	if n1.MoveCommits[node2.Identifier] != hex.EncodeToString(moveCommit2.MoveHash) {
		fmt.Println("Failed to send hash in the opposite direction")
		fmt.Printf("Node 1 [%v]\n", n1.MoveCommits[node1.Identifier])
		fmt.Printf("Hash sent [%v]\n", hex.EncodeToString(moveCommit2.MoveHash))
		t.Fail()
	}

	fmt.Printf("Hash sent [%v]\n", hex.EncodeToString(moveCommit2.MoveHash))

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}