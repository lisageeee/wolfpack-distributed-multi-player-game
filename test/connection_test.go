package test
import (
	"fmt"
	"testing"
	n "../logic/impl"
	"net"
	"../key-helpers"
)

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

// TODO: Need to rethink this - instead of the server keeping track of a history of
// past connections on the global server, let's just store the identifier somewhere on
// the player's computer. We do a check at the beginning to see if an ID has already been
// assigned to them, and if so, then we register them on the server but w/o assigning them
// the id

// Right now, if you register with the same IP address, the global server is like - wut,
// you've already registered, get out of here. This is only if it hasn't deleted this
// address out of its allPlayers list (because it's no longer receiving a heartbeat from it)

func TestNonUniqueID(t *testing.T){
	fmt.Println("Testing that the ID's do not increment if same IP given")
	udp_addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2224")
	pubKey, privKey := key_helpers.GenerateKeys()
	node := n.CreateNodeCommInterface(pubKey, privKey)
	node.LocalAddr = udp_addr1

	udp_addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2224")
	pubKey, privKey = key_helpers.GenerateKeys()
	node2 := n.CreateNodeCommInterface(pubKey, privKey)
	node2.LocalAddr = udp_addr2
	res1 := node.ServerRegister()
	res2 := node2.ServerRegister()

	if res1 != res2 {
		t.FailNow()
	}
	fmt.Println("passed")
}