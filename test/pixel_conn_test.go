package test

import (
	"testing"
	l "../logic/impl"
	p "../pixel/impl"
	"context"
	"time"
	"os/exec"
	"../shared"
	"fmt"
	"syscall"
	key "../key-helpers"
	"regexp"
)

// Reference for killing exec.Command processes + childen:
// https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
var serverStart *exec.Cmd
// This test will fail if you make a breaking change that keeps pixel.go from running
// Inspiration: the breaking change I added that prevented pixel.go from running (wrong image path)
func TestPixelNodeCanRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart = exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started

	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	n := l.CreatePlayerNode(":12500", ":12501",  pub, priv, ":8081")
	go n.RunGame(":12501")

	time.Sleep(2*time.Second) // wait playernode to start

	pixelStart := exec.Command("go", "run", "pixel.go", ":12501")
	pixelStart.Dir = "../pixel"
	pixelStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var err error

	// Start the pixel node, set err to the error returned if any
	go func() {
		_, err = pixelStart.Output()
	}()

	// Wait 5 seconds for errors to return
	time.Sleep(5 * time.Second)

	// If pixel can't start, will get err on this line
	if err != nil {
		fmt.Println("Pixel couldn't start, error:", err)
		panic(err)
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
	// Note: you can close the pixel window after this test finishes (sorry, killing it crashes the next test)
}

// Tests that the logic node is able to send messages to the pixel node
// NOTE: this test sometimes fails unless you run it alone (wooo back to that again)
func TestLogicNodeToPixelComm(t *testing.T) {
	// Start running server
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart = exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started

	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	n := l.CreatePlayerNode(":12300", ":12301", pub, priv, ":8081")
	go n.RunGame(":12301")
	remote := n.GetPixelInterface()

	time.Sleep(1*time.Second) // wait playernode to start

	//Run pixel node
	pixel := p.CreatePixelNode(":12301")
	go pixel.RunRemoteNodeListener()

	// Create a gamestate
	playerLocs := make(map[string]shared.Coord)
	playerLocs[n.Identifier] = shared.Coord{2,1}
	playerLocs["prey"] = shared.Coord{0,1}
	playerMap := shared.PlayerLockMap{Data : playerLocs}
	gameState := shared.GameState{
		PlayerLocs: playerMap,
	}

	// Send from the remote interface to the pixel node
	remote.SendPlayerGameState(gameState)

	// Read from the pixel node's channel
	pixelGameState := <- pixel.NewGameStates

	// Check it was sent correctly
	if pixelGameState.PlayerLoc.X != gameState.PlayerLocs.Data[n.Identifier].X {
		t.Fail()
	}
	if pixelGameState.PlayerLoc.Y != gameState.PlayerLocs.Data[n.Identifier].Y {
		t.Fail()
	}
	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()

}

// Tests that the pixel node can send messages to the logic node
// NOTE: this test sometimes fails unless you run it alone (wooo back to that again)
func TestPixelNodeMove(t *testing.T) {
	// Start running server
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	serverStart = exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.Dir = "../server"
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Start()

	time.Sleep(3 * time.Second) // wait for server to get started

	// Create player node, run it and get pixel interface
	pub, priv := key.GenerateKeys()
	n := l.CreatePlayerNode(":12303", ":12304", pub, priv, ":8081")
	go n.RunGame(":12304")
	loc := n.GameState.PlayerLocs.Data[n.Identifier] // get the initial game render state

	time.Sleep(1*time.Second) // wait playernode to start

	// Start pixel node
	pixel := p.CreatePixelNode(":12304")
	go pixel.RunRemoteNodeListener()

	pixel.SendMove("up")

	// Wait a tick for the move to be sent
	time.Sleep(100*time.Millisecond)

	// Check that the player has moved up
	newState := n.GameState

	fmt.Println(newState, loc)
	if newState.PlayerLocs.Data[n.Identifier].X != loc.X {
		fmt.Println("Player x changed when it shouldn't have")
		t.Fail()
	}
	if newState.PlayerLocs.Data[n.Identifier].Y - 1 != loc.Y {
		fmt.Println("Player Y didn't change when it should have")
		t.Fail()
	}

	// Reset to try moving down
	loc = newState.PlayerLocs.Data[n.Identifier]
	pixel.SendMove("down")

	// Wait a tick for the move to be sent
	time.Sleep(100*time.Millisecond)

	// Check that the player has moved down
	newState = n.GameState
	fmt.Println(newState, loc)
	if newState.PlayerLocs.Data[n.Identifier].X != loc.X {
		fmt.Println("Player x changed when it shouldn't have")
		t.Fail()
	}
	if newState.PlayerLocs.Data[n.Identifier].Y != loc.Y - 1 {
		fmt.Println("Player Y didn't change when it should have")
		t.Fail()
	}
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
}

func TestPixelNodeGetConfigFromServer(t *testing.T) {
	// Start running server
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go", "9090", "1") // run with the alt gamestate
	serverStart.Dir = "../server"
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Start()

	time.Sleep(3 * time.Second) // wait for server to get started

	// Create player node, run it and get pixel interface
	pub, priv := key.GenerateKeys()
	n := l.CreatePlayerNode(":12504", ":12555", pub, priv, ":9090")
	go n.RunGame(":12555")

	time.Sleep(2*time.Second) // wait playernode to start

	// Check the playernode has the alternate config (600x600)
	if n.GameConfig.Settings.WindowsY != 600 || n.GameConfig.Settings.WindowsX != 600 {
		t.Fail()
	}

	// Start pixel node
	pixel := p.CreatePixelNode(":12555")
	go pixel.RunRemoteNodeListener()

	// Check the pixel node has the alternate config
	if pixel.Geom.GetX() != 600 || pixel.Geom.GetY() != 600 {
		t.Fail()
	}

	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
}

func TestPixelSortsScores(t *testing.T) {
	fakeScoreMap := make(map[string]int)
	fakeScoreMap["loser2"] = 100
	fakeScoreMap["loser1"] = 300
	fakeScoreMap["loser3"] = 50
	fakeScoreMap["winner"] = 1000

	scoreString := p.SortScores(fakeScoreMap)

	winnerRegex, _ := regexp.Compile("winner")
	secondRegex, _ := regexp.Compile("loser1")
	thirdRegex, _ := regexp.Compile("loser2")
	lastRegex, _ := regexp.Compile("loser3")

	winner := winnerRegex.FindIndex([]byte(scoreString))[0]
	second := secondRegex.FindIndex([]byte(scoreString))[0]
	third := thirdRegex.FindIndex([]byte(scoreString))[0]
	last := lastRegex.FindIndex([]byte(scoreString))[0]

	if winner > second || winner > third || winner > last {
		fmt.Println("Score order incorrect, fail - winner is not first")
		t.Fail()
	}

	if second > third || second > last {
		fmt.Println("Score order incorrect, fail - second is not second")
		t.Fail()
	}

	if third > last {
		fmt.Println("Score order incorrect, fail - third is after last")
		t.Fail()
	}
}
