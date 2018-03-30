package solitary


import (
"testing"
"time"
"os/exec"
"syscall"
key "../key-helpers"
l "../logic/impl"
"../shared"
"context"
p "../pixel/impl"
)

func TestRenderMoveOnIncoming(t *testing.T) {
	const serverPort= "9081"
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go", serverPort)
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(4 * time.Second) // wait for server to get started

	// Create player node, run it and get pixel interface
	pub, priv := key.GenerateKeys()
	node1 := l.CreatePlayerNode(":2830", ":2831", pub, priv, ":9081")
	go node1.RunGame(":2831")

	pub, priv = key.GenerateKeys()
	node2 := l.CreatePlayerNode("3830", ":3831", pub, priv, ":9081")
	n2 := node2.GetNodeInterface()

	time.Sleep(2 * time.Second) // wait playernode to start

	// Start pixel node
	pixel := p.CreatePixelNode(":2831")
	go pixel.RunRemoteNodeListener()

	time.Sleep(1 * time.Second)
	testCoord := shared.Coord{6, 3}
	n2.SendMoveToNodes(&testCoord)
}
