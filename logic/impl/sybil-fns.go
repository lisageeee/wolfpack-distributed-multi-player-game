package impl

import (
	"../../shared"
)

func (pn * PlayerNode) SybilGhostMove(move string) (shared.Coord, bool) {
	// Get current player state
	pn.GameState.PlayerLocs.RLock()
	playerLoc := pn.GameState.PlayerLocs.Data[pn.Identifier]
	pn.GameState.PlayerLocs.RUnlock()

	// Calculate new position with move
	newPosition := shared.Coord{X: playerLoc.X, Y: playerLoc.Y}
	switch move {
	case "up":
		newPosition.Y = newPosition.Y + 1
	case "down":
		newPosition.Y = newPosition.Y - 1
	case "left":
		newPosition.X = newPosition.X - 1
	case "right":
		newPosition.X = newPosition.X + 1
	}

	pn.GameState.PlayerLocs.Lock()
	pn.GameState.PlayerLocs.Data[pn.Identifier] = newPosition
	pn.GameState.PlayerLocs.Unlock()

	return newPosition, true
}

// Takes in a new coordinate for this node and sends it to all other nodes.
func(n* NodeCommInterface) SendGhostMoveToNodes(move *shared.Coord){
	if move == nil {
		return
	}

	sequenceNumber++
	moveId := n.CreateMove(move)
	message := NodeMessage{
		MessageType: "move",
		Identifier:  n.PlayerNode.Identifier,
		Move:        moveId,
		Addr:        n.LocalAddr.String(),
		Seq:         sequenceNumber,
	}

	toSend := sendMessage(n.Log, message, "Sendin' ghost move")
	n.MessagesToSend <- &PendingMessage{Recipient: "all", Message: toSend}
}