package solitary_test

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
	"path/filepath"
)
func DeleteVectorFile(){
	names, _ := filepath.Glob("*Log.txt")
	for _, val := range names{
		_, err := os.Stat(val)
		if err == nil {
			os.Remove(val)
		}

	}
}
func TestGoVectorClockTick(t *testing.T) {
	DeleteVectorFile()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()
	time.Sleep(3 * time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":13900", ":13901", pub, priv, ":8081")
	n1 := node1.GetNodeInterface()
	if n1.Log.GetCurrentVCAsClock().LastUpdate() != 1 {
		fmt.Println("Clock is farked")
		t.FailNow()
	}
	n1.HeartAttack <- true
	time.Sleep(time.Second*8)
}
func TestGoVectorSingleMessage(t *testing.T) {
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
	node1 := l.CreatePlayerNode(":12800", ":12801", pub, priv, ":8081")
	n1 := node1.GetNodeInterface()
	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":12900", ":12901", pub, priv, ":8081")
	n2 := node2.GetNodeInterface()
	time.Sleep(time.Second)
	// Check nodes are connected to each other
	if len(n2.OtherNodes) != 1 ||  len(n1.OtherNodes) != 1 {
		fmt.Println("Nodes do not have a mutual connection, fail")
		t.Fail()
	}
	if n1.Log.GetCurrentVCAsClock().LastUpdate()!= 2 || n2.Log.GetCurrentVCAsClock().LastUpdate()!= 2{
		fmt.Println(n1.Log.GetCurrentVCAsClock())
		fmt.Println(n2.Log.GetCurrentVCAsClock())
		fmt.Println("Clock is farked")
		t.FailNow()
	}
	// Test sending a move from one node to another
	testCoord := shared.Coord{7,7}
	n1.SendMoveToNodes(&testCoord)

	time.Sleep(100*time.Millisecond)

	if n2.PlayerNode.GameState.PlayerLocs.Data[node1.Identifier] != testCoord {
		fmt.Println("Failed to send testCoord from Node 1 to node 2")
		t.Fail()
	}
	if n1.Log.GetCurrentVCAsClock().LastUpdate()!= 3 || n2.Log.GetCurrentVCAsClock().LastUpdate()!= 3{
		fmt.Println(n2.Log.GetCurrentVCAsClock())
		fmt.Println("Clock is farked")
		t.FailNow()
	}

	n1.HeartAttack<- true
	n2.HeartAttack<- true
	time.Sleep(time.Second*8)
	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestGoVectorDoubleMessage(t *testing.T) {
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
	node1 := l.CreatePlayerNode(":14800", ":14801", pub, priv, ":8081")
	n1 := node1.GetNodeInterface()
	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":14900", ":14901", pub, priv, ":8081")
	n2 := node2.GetNodeInterface()
	time.Sleep(time.Second)
	// Check nodes are connected to each other
	if len(n2.OtherNodes) != 1 ||  len(n1.OtherNodes) != 1 {
		fmt.Println("Nodes do not have a mutual connection, fail")
		t.Fail()
	}
	if n1.Log.GetCurrentVCAsClock().LastUpdate()!= 2 || n2.Log.GetCurrentVCAsClock().LastUpdate()!= 2{
		fmt.Println(n1.Log.GetCurrentVCAsClock())
		fmt.Println(n2.Log.GetCurrentVCAsClock())
		fmt.Println("Clock is farked")
		t.FailNow()
	}
	// Test sending a move from one node to another
	testCoord := shared.Coord{7,7}
	n1.SendMoveToNodes(&testCoord)

	time.Sleep(100*time.Millisecond)

	if n2.PlayerNode.GameState.PlayerLocs.Data[node1.Identifier] != testCoord {
		fmt.Println("Failed to send testCoord from Node 1 to node 2")
		t.Fail()
	}

	n2.SendMoveToNodes(&testCoord)
	time.Sleep(100*time.Millisecond)
	if n1.Log.GetCurrentVCAsClock().LastUpdate()!= 4 || n2.Log.GetCurrentVCAsClock().LastUpdate()!= 4{
		fmt.Println(n2.Log.GetCurrentVCAsClock())
		fmt.Println("Clock is farked")
		t.FailNow()
	}
	n1.HeartAttack<- true
	n2.HeartAttack<- true
	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}