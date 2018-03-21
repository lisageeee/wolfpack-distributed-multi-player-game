package test
import (
	"fmt"
	"testing"
	n "../logic/impl"
	"net"
	"../key-helpers"
	"time"
)

func TestHeartbeat(t *testing.T) {
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

	fmt.Printf("[%v]\n", node3.Connections)
	if len(node3.Connections) != 2 {
		t.FailNow()
	}

	time.Sleep(5*time.Second)

	node3.GetNodes()
	fmt.Printf("[%v]\n", node3.Connections)
	if len(node3.Connections) != 1 {
		t.FailNow()
	}

	fmt.Printf("TEST PASSED: Node3 has the connection details of Node1 only\n")
}

func TestIncrementingID(t *testing.T){
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
		t.FailNow()
	}

	fmt.Printf("TEST PASSED: Res1's id is [%s] and Res2's id is [%s]\n", res1, res2)
}

// TODO: Test node re-joins and has been assigned an identifier - cannot assume that it's
// connecting from the same IP address

// Right now, if you register with the same IP address, the global server is like - wut,
// you've already registered, get out of here. This is only if it hasn't deleted this
// address out of its allPlayers list (because it's no longer receiving a heartbeat from it)