package test

import (
	"os/exec"
	"testing"
	"context"
	"time"
	"syscall"
	"fmt"
	key "../key-helpers"
	l "../logic/impl"
)

// This test will fail if you make a breaking change that keeps pixel.go from running
// Inspiration: the breaking change I added that prevented pixel.go from running (wrong image path)
func TestNodeToNodeComm(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(3*time.Second) // wait for server to get started
	// Create player node and get pixel interface
	pub, priv := key.GenerateKeys()
	_ = l.CreatePlayerNode(":12400", ":12401", ":12402", pub, priv)

	pixelStart := exec.Command("go", "run", "pixel.go", ":12401", ":12402")
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
		t.Fail()
	}

	// Kill after done + all children
	syscall.Kill(-pixelStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()

}
