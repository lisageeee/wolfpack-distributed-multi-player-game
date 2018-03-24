package test
import (
	"fmt"
	"testing"
	n "../logic/impl"
	"net"
	"../key-helpers"
	"time"
	"context"
	"os/exec"
	"syscall"
)

func TestHeartbeat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(2 * time.Second) // give server time to start

	fmt.Println("Testing the heartbeat functionality")
	udp_addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2124")
	pubKey, privKey := key_helpers.GenerateKeys()
	node := n.CreateNodeCommInterface(pubKey, privKey)
	node.LocalAddr = udp_addr1
	_ = node.ServerRegister()
	go node.SendHeartbeat()

	udp_addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2125")
	pubKey, privKey = key_helpers.GenerateKeys()
	node2 := n.CreateNodeCommInterface(pubKey, privKey)
	node2.LocalAddr = udp_addr2
	_ = node2.ServerRegister()

	udp_addr3, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2126")
	pubKey, privKey = key_helpers.GenerateKeys()
	node3 := n.CreateNodeCommInterface(pubKey, privKey)
	node3.LocalAddr = udp_addr3
	_ = node3.ServerRegister()
	go node3.SendHeartbeat()

	fmt.Printf("[%v]\n", node3.OtherNodes)
	if len(node3.OtherNodes) != 2 {
		t.Fail()
	}

	time.Sleep(5*time.Second)

	node3.GetNodes()
	fmt.Printf("[%v]\n", node3.OtherNodes)
	if len(node3.OtherNodes) != 1 {
		t.Fail()
	}

	fmt.Printf("TEST PASSED: Node3 has the connection details of Node1 only\n")

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestIncrementingID(t *testing.T){
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(2 * time.Second) // give server time to start

	fmt.Println("Testing that the ID's increment")
	udp_addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2123")
	pubKey, privKey := key_helpers.GenerateKeys()
	node := n.CreateNodeCommInterface(pubKey, privKey)
	node.LocalAddr = udp_addr1
	res1 := node.ServerRegister()

	udp_addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2023")
	pubKey, privKey = key_helpers.GenerateKeys()
	node2 := n.CreateNodeCommInterface(pubKey, privKey)
	node2.LocalAddr = udp_addr2
	res2 := node2.ServerRegister()

	if res1 == res2 {
		t.Fail()
	}

	fmt.Printf("TEST PASSED: Res1's id is [%s] and Res2's id is [%s]\n", res1, res2)

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

// TODO: Test node re-joins and has been assigned an identifier - cannot assume that it's
// connecting from the same IP address

// Right now, if you register with the same IP address, the global server is like - wut,
// you've already registered, get out of here. This is only if it hasn't deleted this
// address out of its allPlayers list (because it's no longer receiving a heartbeat from it)