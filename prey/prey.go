//package prey
//
//import (
//	"../geometry"
//	"github.com/faiface/pixel"
//)
//
//type PreyRunner struct {
//	geo geometry.PixelManager
//	position pixel.Vec
//}
//
//func (pr *PreyRunner) GetPosition() (pixel.Vec) {
//	return pr.position
//}
//
//func (pr *PreyRunner) Move() (pixel.Vec) {
//	//random := rand.Float64()
//	//X := pr.position.X
//	//Y:= pr.position.Y
//	//step := pr.geo.GetSpriteSize()
//	//var newVec pixel.Vec
//	//switch {
//	//	case random < 0.25:
//	//		newVec = pixel.V(X - step, Y + step)
//	//	case random < 0.5:
//	//		newVec = pixel.V(X + step, Y + step)
//	//	case random < 0.75:
//	//		newVec = pixel.V(X + step, Y - step)
//	//	default:
//	//		newVec = pixel.V(X - step, Y - step)
//	//}
//	return pr.position
//}

package main

import (
	"fmt"
	"os"
	_ "image/png"
	_ "image/jpeg"
	logicImpl "./impl"
	"../key-helpers"
)

// Entrypoint for the player (logic) node, creates the node and all interfaces by calling the playerNode constructor
// and calling runGame
func main() {
	fmt.Println("hello world")

	// Default IP addresses if none provided
	nodeListenerAddr := "127.0.0.1:0"
	playerListenerIpAddress := "127.0.0.1:12345"
	serverAddr := ":8081"
	// Can start with an IP as param
	if len(os.Args) > 3 {
		nodeListenerAddr = os.Args[1]
		playerListenerIpAddress = os.Args[2]
		serverAddr = os.Args[3]
	} else if len(os.Args) > 2 {
		nodeListenerAddr = os.Args[1]
		playerListenerIpAddress = os.Args[2]
	} else if len(os.Args)>1{
		nodeListenerAddr = "127.0.0.1:0"
		playerListenerIpAddress = os.Args[1]
	}

	pubKey, privKey := key_helpers.GenerateKeys()
	node := logicImpl.CreatePreyNode(nodeListenerAddr, playerListenerIpAddress, pubKey, privKey, serverAddr)
	node.RunGame(playerListenerIpAddress)
}