package test
import (
	"fmt"
	"testing"
	n "../logic/impl"
	"net"
)


func TestIncrementingID(t *testing.T){
	fmt.Println("Testing that the ID's increment")
	udp_addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2123")
	node := n.CreateNodeCommInterface()
	node.Address = udp_addr1

	udp_addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2223")
	node2 := n.CreateNodeCommInterface()
	node2.Address = udp_addr2
	res1 := node.ServerRegister()
	res2 := node2.ServerRegister()

	if res1.Identifier-1 != res2.Identifier{
		t.FailNow()
	}
	fmt.Println("passed")
}
func TestNonUniqueID(t *testing.T){
	fmt.Println("Testing that the ID's do not increment if same IP given")
	udp_addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2223")
	node := n.CreateNodeCommInterface()
	node.Address = udp_addr1

	udp_addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2223")
	node2 := n.CreateNodeCommInterface()
	node2.Address = udp_addr2
	res1 := node.ServerRegister()
	res2 := node2.ServerRegister()

	if res1.Identifier != res2.Identifier {
		t.FailNow()
	}
	fmt.Println("passed")
}