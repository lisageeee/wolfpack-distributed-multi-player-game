package solitary

import (
	"time"
	"fmt"
	"syscall"
	"testing"
	"os/exec"
	key "../key-helpers"
	"context"
	l "../logic/impl"
	"../shared"
)

func TestNodeToNodeSendingGamestateOnInit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":16030", ":16031", pub, priv, ":8081")

	//n1 := node1.GetNodeInterface()

	node1.GameState.PlayerLocs.Data["prey"] = shared.Coord{6, 5}
	node1.GameState.PlayerLocs.Data[node1.Identifier] = shared.Coord{6,6}

	time.Sleep(1*time.Second)


	go node1.RunGame(":16031")
	channel := node1.GetPlayerCommChannel()
	channel <- "down"

	time.Sleep(200*time.Millisecond)

	// Check that we have captured the prey now
	if node1.GameState.PlayerScores.Data[node1.Identifier] != node1.GameConfig.CatchWorth {
		fmt.Println("Did not score, fail")
	}

	channel <- "up"
	time.Sleep(200*time.Millisecond)

	channel <- "down"
	time.Sleep(200*time.Millisecond)

	// Check that we have captured the prey now
	if node1.GameState.PlayerScores.Data[node1.Identifier] != node1.GameConfig.CatchWorth * 2 {
		fmt.Println("Did not score again, fail")
	}

	time.Sleep(1*time.Second)

	// Create a second node
	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode(":19730", ":19731", pub, priv, ":8081")
	go node2.RunGame(":19731")

	time.Sleep(2*time.Second)

	// Check if node2 has node1's score
	_, ok := node2.GameState.PlayerScores.Data[node1.Identifier]
	if !ok {
		fmt.Println("Node 2 doesn't have a score for node 1, fail")
		t.Fail()
	} else {
		if node2.GameState.PlayerScores.Data[node1.Identifier] != node2.GameConfig.CatchWorth * 2 {
			fmt.Printf("Node 2 doesn't have the correct score, expected %v, got %v",
				node2.GameConfig.CatchWorth * 2, node2.GameState.PlayerScores.Data[node1.Identifier])
			t.Fail()
		}
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}
