package test

import (
	"testing"
	l "../logic/impl"
	"context"
	"time"
	"os/exec"
)

//NOTE:
//nodeListenerAddr = os.Args[1]
//playerListenerIpAddress = os.Args[2]
//pixelIpAddress = os.Args[3]

func TestPassCLI(t *testing.T) {
	// Start running server
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.Dir = "../server"
	serverStart.Start()

	// Create player node
	l.CreatePlayerNode(":12300", ":12301", ":12302")
	// l.CreatePlayerNode(":12303", ":12304", ":12305")

	//Run pixel interface
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel2()
	pixelStart := exec.CommandContext(ctx2, "go", "run", "pixel-node.go", ":12301", ":12302")
	pixelStart.Dir = ".."
	pixelStart.Start()

}