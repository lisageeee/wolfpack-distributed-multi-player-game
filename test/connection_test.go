package test
import (
	"fmt"
	"testing"
	n ".."
)


func TestIncrementingID(t *testing.T){
	fmt.Println("Testing that the ID's increment")
	resp1 := n.ServerRegister(":0")
	resp2 := n.ServerRegister(":1")
	if resp2.Identifier-1 != resp1.Identifier{
		t.FailNow()
	}
	fmt.Println("passed")
}
func TestNonUniqueID(t *testing.T){
	fmt.Println("Testing that the ID's do not increment if same IP given")
	resp1 := n.ServerRegister(":0")
	resp2 := n.ServerRegister(":0")
	if resp2.Identifier != resp1.Identifier{
		t.FailNow()
	}
	fmt.Println("passed")
}