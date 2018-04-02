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
	"os"
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
	pubKey, privKey := key_helpers.GenerateKeys()
	node := n.CreateNodeCommInterface(pubKey, privKey, ":8081")
	go node.ManageOtherNodes()

	addr1, nodeConn := n.StartListenerUDP(":2124")
	node.LocalAddr = addr1
	go node.RunListener(nodeConn, addr1.String())

	_ = node.ServerRegister()
	go node.SendHeartbeat()

	pubKey, privKey = key_helpers.GenerateKeys()
	node2 := n.CreateNodeCommInterface(pubKey, privKey, ":8081")
	go node2.ManageOtherNodes()

	addr2, nodeConn2 := n.StartListenerUDP(":2125")
	node2.LocalAddr = addr2
	go node2.RunListener(nodeConn2, addr2.String())
	_ = node2.ServerRegister()


	udp_addr3, _ := net.ResolveUDPAddr("udp", ":2126")
	pubKey, privKey = key_helpers.GenerateKeys()
	node3 := n.CreateNodeCommInterface(pubKey, privKey, ":8081")
	go node3.ManageOtherNodes()

	node3.LocalAddr = udp_addr3
	_ = node3.ServerRegister()
	go node3.SendHeartbeat()

	fmt.Printf("[%v]\n", node3.OtherNodes)
	if len(node3.OtherNodes) != 2 {
		fmt.Println("Fail, expected node3 to have 2 connections, has ", len(node3.OtherNodes))
		t.Fail()
	}

	time.Sleep(5*time.Second)

	// This part of the test is incorrect now that getNodes won't wholesale replace the other nodes
	//node3.GetNodes()
	//fmt.Printf("[%v]\n", node3.OtherNodes)
	//if len(node3.OtherNodes) != 1 {
	//	fmt.Println("Fail, expected node3 to have 1 connections, has ", len(node3.OtherNodes))
	//	t.Fail()
	//} else {
	//	fmt.Printf("TEST PASSED: Node3 has the connection details of Node1 only\n")
	//}

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
	udp_addr1, _ := net.ResolveUDPAddr("udp", ":2123")
	pubKey, privKey := key_helpers.GenerateKeys()
	node := n.CreateNodeCommInterface(pubKey, privKey, ":8081")
	node.LocalAddr = udp_addr1
	res1 := node.ServerRegister()

	udp_addr2, _ := net.ResolveUDPAddr("udp", ":2023")
	pubKey, privKey = key_helpers.GenerateKeys()
	node2 := n.CreateNodeCommInterface(pubKey, privKey, ":8081")
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

func TestServerCommandLineArgs(t *testing.T) {
	const serverPort = "9001"
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go", serverPort)
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(4 * time.Second) // give server time to start

	fmt.Println("Testing that the ID's increment")
	udp_addr1, _ := net.ResolveUDPAddr("udp", ":2123")
	pubKey, privKey := key_helpers.GenerateKeys()
	node := n.CreateNodeCommInterface(pubKey, privKey, ":" +serverPort)
	node.LocalAddr = udp_addr1
	res1 := node.ServerRegister()

	udp_addr2, _ := net.ResolveUDPAddr("udp", ":2023")
	pubKey, privKey = key_helpers.GenerateKeys()
	node2 := n.CreateNodeCommInterface(pubKey, privKey, ":"+serverPort)
	node2.LocalAddr = udp_addr2
	res2 := node2.ServerRegister()

	if res1 == res2 {
		t.Fail()
	}

	fmt.Printf("TEST PASSED: Nodes [%s] and [%s] able to connect to server running at port [%s]\n", res1, res2, serverPort)

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

func TestServerDies(t *testing.T) {
	const serverPort = "8008"
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go", serverPort)
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(4 * time.Second) // give server time to start

	fmt.Println("Testing that the ID's increment")
	udp_addr1, _ := net.ResolveUDPAddr("udp", ":2123")
	pubKey, privKey := key_helpers.GenerateKeys()
	node := n.CreateNodeCommInterface(pubKey, privKey, ":" +serverPort)
	node.LocalAddr = udp_addr1
	res1 := node.ServerRegister()

	udp_addr2, _ := net.ResolveUDPAddr("udp", ":2023")
	pubKey, privKey = key_helpers.GenerateKeys()
	node2 := n.CreateNodeCommInterface(pubKey, privKey, ":"+serverPort)
	node2.LocalAddr = udp_addr2
	res2 := node2.ServerRegister()
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
	if res1 == res2 {
		t.Fail()
	} else {
		fmt.Printf("TEST PASSED: Nodes [%s] and [%s] able to connect to server running at port [%s]\n", res1, res2, serverPort)
	}

	time.Sleep(time.Second)

	// Test if still alive

	var _ignored bool
	err := node.ServerConn.Call("GServer.Heartbeat", *node.PubKey, &_ignored)
	if err == nil {
		fmt.Println("Server should be dead")
		os.Exit(1)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel2()
	serverStart2 := exec.CommandContext(ctx2, "go", "run", "server.go", serverPort)
	serverStart2.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart2.Dir = "../server"
	server_err := serverStart2.Start()
	if server_err!= nil {
		fmt.Println(server_err)
	}
	time.Sleep(3*time.Second)
	node.Reregister()
	err = node.ServerConn.Call("GServer.Heartbeat", *node.PubKey, &_ignored)
	if err != nil {
		fmt.Println("Server should be alive" )
		os.Exit(1)
	}

	syscall.Kill(-serverStart2.Process.Pid, syscall.SIGKILL)
	serverStart2.Process.Kill()
}

func TestPubKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7 * time.Second)
	defer cancel()
	serverStart := exec.CommandContext(ctx, "go", "run", "server.go")
	serverStart.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	serverStart.Dir = "../server"
	serverStart.Start()

	time.Sleep(2 * time.Second) // give server time to start

	fmt.Println("Testing public key transmission")
	pubKey1, privKey1 := key_helpers.GenerateKeys()
	node := n.CreateNodeCommInterface(pubKey1, privKey1, ":8081")
	go node.ManageOtherNodes()

	addr1, nodeConn := n.StartListenerUDP(":2140")
	node.LocalAddr = addr1
	go node.RunListener(nodeConn, addr1.String())

	nodeId := node.ServerRegister()
	go node.SendHeartbeat()

	pubKey2, privKey2 := key_helpers.GenerateKeys()
	node2 := n.CreateNodeCommInterface(pubKey2, privKey2, ":8081")
	go node2.ManageOtherNodes()

	addr2, nodeConn2 := n.StartListenerUDP(":2150")
	node2.LocalAddr = addr2
	go node2.RunListener(nodeConn2, addr2.String())
	_ = node2.ServerRegister()

	time.Sleep(2*time.Second)
	fmt.Println(node2.NodeKeys, pubKey1, "THGJFLDKGJF", nodeId)

	if pubKey1.X.Cmp(node2.NodeKeys[nodeId].X) != 0 {
		fmt.Printf("Fail, node 1 does not have node 1's public key")
	}

	if pubKey1.Y.Cmp(node2.NodeKeys[nodeId].Y) != 0 {
		fmt.Printf("Fail, node 1 does not have node 1's public key")
	}

	if pubKey1.Curve != node2.NodeKeys[nodeId].Curve {
		fmt.Printf("Fail, node 1 does not have node 1's public key")
	}

	// Kill after done + all children
	syscall.Kill(-serverStart.Process.Pid, syscall.SIGKILL)
	serverStart.Process.Kill()
}

// TODO: Test node re-joins and has been assigned an identifier - cannot assume that it's
// connecting from the same IP address

// Right now, if you register with the same IP address, the global server is like - wut,
// you've already registered, get out of here. This is only if it hasn't deleted this
// address out of its allPlayers list (because it's no longer receiving a heartbeat from it)