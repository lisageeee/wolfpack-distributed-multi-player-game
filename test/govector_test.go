package test

import (
	"time"
	"os/exec"
	"syscall"
	"fmt"
	"testing"
	"context"
	key "../key-helpers"
	l "../logic/impl"
	"../shared"
	"os"
)
func DeleteVectorFile(){
	_, err := os.Stat("LogicNodeFile-Log.txt")
	if err == nil {
		os.Remove("LogicNodeFile-Log.txt")
	}
}
func TestGoVectorClockTick(t *testing.T) {
	DeleteVectorFile()
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":12800", ":12801", ":12802", pub, priv)

	n1 := node1.GetNodeInterface()

	if n1.Log.GetCurrentVCAsClock().LastUpdate()!= 1{
		fmt.Println("Clock is farked")
		t.FailNow()
	}
	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":12900", ":12901", ":12092", pub, priv)
	n2 := node2.GetNodeInterface()



	// Check nodes are connected to each other
	if len(n2.OtherNodes) != 1 || len(n1.OtherNodes) != 1{
		fmt.Println("Nodes do not have a mutual connection, fail")
		t.Fail()
	}

	// Test sending a move from one node to another
	testCoord := shared.Coord{7,7}
	n1.SendMoveToNodes(&testCoord)
	time.Sleep(100*time.Millisecond)

	if n2.PlayerNode.GameState.PlayerLocs[node1.Identifier] != testCoord {
		fmt.Println("Failed to send testCoord from Node 1 to node 2")
		t.Fail()
	}

	testCoord = shared.Coord{6,3}
	n2.SendMoveToNodes(&testCoord)
	time.Sleep(100*time.Millisecond)

	if n1.PlayerNode.GameState.PlayerLocs[node2.Identifier] != testCoord {
		fmt.Println("Failed to send testCoord in the opposite direction")
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}
