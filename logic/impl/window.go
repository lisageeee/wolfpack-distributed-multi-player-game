package impl

import (
	"../../shared"
	"sync"
	"reflect"
	"../../geometry"
)

type MoveSeq struct{
	seq 	uint64
	coords	*shared.Coord
}
const NUMMOVESTOKEEP = 10
type RunningWindow struct{
	sync.Mutex
	Map map[string][NUMMOVESTOKEEP]MoveSeq
	PreySeq uint64
}

func(rw *RunningWindow)Add(id string, seq uint64, coords *shared.Coord){
	rw.Lock()
	defer rw.Unlock()
	if _, ok := rw.Map[id]; !ok{
		rw.Map[id] = [NUMMOVESTOKEEP]MoveSeq{}
	}
	if id == "prey"{
		rw.PreySeq = seq
	}
	movSeq := rw.Map[id]
	for i := NUMMOVESTOKEEP-1; i>0; i--{
		movSeq[i] = movSeq[i-1]
	}

	movSeq[0] = MoveSeq{seq, coords}
	rw.Map[id] = movSeq

}

func(rw *RunningWindow)Match(id string, seq uint64, coords * shared.Coord)bool{
	rw.Lock()
	defer rw.Unlock()
	pastMoves := rw.Map[id]
	for i := 0; i < NUMMOVESTOKEEP; i++{
		if pastMoves[i].seq == seq{
			return reflect.DeepEqual(pastMoves[i].coords, coords)
		}
	}

	return false
}

func (rw *RunningWindow) IsNotTeleport (id string, seq uint64, coords * shared.Coord,
	manager *geometry.GridManager) (bool) {
	rw.Lock()
	defer rw.Unlock()
	_, ok := rw.Map[id]
	if !ok {
		return true
	}
	pastMoves := rw.Map[id]

	// Track the highest sequence number we have seen and its index
	var highestSeq uint64 = 0
	index := 0

	// Look for the sequence number before this
	for i := 0; i < NUMMOVESTOKEEP; i++{
		// Keep track of the most recent move
		if pastMoves[i].seq > highestSeq {
			highestSeq = pastMoves[i].seq
			index = i
		}
		if pastMoves[i].seq == seq - 1 {
			return manager.IsNotTeleporting(*pastMoves[i].coords, *coords)
		}
	}

	// If we didn't store a sequence number before this, look for the most recent one to compare
	return manager.IsNotTeleporting(*pastMoves[index].coords, *coords)
}