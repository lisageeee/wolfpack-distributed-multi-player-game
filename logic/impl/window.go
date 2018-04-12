package impl

import (
	"../../shared"
	"sync"
	"reflect"
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
	for i := 0; i<3; i++{
		if pastMoves[i].seq == seq{
			return reflect.DeepEqual(pastMoves[i].coords, coords)
		}
	}

	return false
}