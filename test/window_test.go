package test
import (
	l "../logic/impl"
	"testing"
	"../shared"
	"fmt"
	"time"
	"../geometry"
)

func before()l.RunningWindow {
	return l.RunningWindow{Map:make(map[string][l.NUMMOVESTOKEEP]l.MoveSeq)}
}


func TestAddOne(t *testing.T) {
	rw := before()
	fmt.Println("Add One Test")
	t1 := time.Now()
	rw.Add("id1", uint64(1), &shared.Coord{10, 6})
	fmt.Println("Elapsed Time: ", time.Now().Sub(t1).Seconds())
	if !rw.Match("id1", uint64(1), &shared.Coord{10, 6}){
		t.FailNow()
	}
	fmt.Println("Passed")
}



func TestAddFive(t *testing.T) {
	rw := before()
	fmt.Println("Add Five Test")

	for i := 1; i < 6; i++ {
		rw.Add("id1", uint64(i), &shared.Coord{10,
			6})
	}
	for i := 3; i < 6; i++ {
		if !rw.Match("id1", uint64(i), &shared.Coord{10, 6}) {
			t.FailNow()
		}
	}
	if rw.Match("id1", uint64(1), &shared.Coord{10, 6}) {
		t.FailNow()
	}
	fmt.Println("Passed")
}

func TestAddFiveMultipleIds(t *testing.T) {
	rw := before()
	fmt.Println("Add Five Test Multiple Ids")

	for i := 1; i < 6; i++ {
		rw.Add("id1", uint64(i), &shared.Coord{i+4,
			i-2})
	}
	for i := 1; i < 6; i++ {
		rw.Add("id2", uint64(i), &shared.Coord{i+4,
			i-2})
	}
	for i := 3; i < 6; i++ {
		if !rw.Match("id1", uint64(i), &shared.Coord{i+4, i-2}) {
			t.FailNow()
		}
	}
	for i := 3; i < 6; i++ {
		if !rw.Match("id2", uint64(i), &shared.Coord{i+4, i-2}) {
			t.FailNow()
		}
	}
	fmt.Println("Passed")
}
func TestAddGoRoutines(t *testing.T) {
	rw := before()
	fmt.Println("Add Test Goroutines")
	// This fails sometimes, does it really take >500 ms to add to this map? No?!

	go func() {
		for i := 1; i < 6; i++ {
			rw.Add("id1", uint64(i), &shared.Coord{i + 4,
				i - 2})
		}
	}()
	go func(){
		for i := 1; i < 6; i++ {
			rw.Add("id2", uint64(i), &shared.Coord{i+4,
				i-2})
		}
	}()
	time.Sleep(time.Millisecond*200)
	for i := 3; i < 6; i++ {
		if !rw.Match("id1", uint64(i), &shared.Coord{i+4, i-2}) {
			fmt.Println("id1 ", i)
			t.FailNow()
		}
	}
	for i := 3; i < 6; i++ {
		if !rw.Match("id2", uint64(i), &shared.Coord{i+4, i-2}) {
			fmt.Println("id2 ", i)
			t.FailNow()
		}
	}
	fmt.Println("Passed")
}
func TestPrey(t *testing.T) {
	rw := before()
	fmt.Println("Test Prey")

	for i := 1; i < 6; i++ {
		rw.Add("prey", uint64(i), &shared.Coord{i+4,
			i-2})
	}
	for i := 3; i < 6; i++ {
		if !rw.Match("prey", uint64(i), &shared.Coord{i+4, i-2}) {
			fmt.Println("prey ", i)
			t.FailNow()
		}
	}
	if rw.PreySeq != 5{
		t.FailNow()
	}
	fmt.Println("Passed")
}

func TestNoTeleportInSeq(t *testing.T) {
	rw := before()
	gs := shared.InitialGameSettings{3000, 3000,
		[]shared.Coord{{1,1}, {10, 90}, {23, 99} }, 200}
	gm := geometry.CreateNewGridManager(gs)

	for i := 1; i < 3; i++ {
		rw.Add("1", uint64(i), &shared.Coord{2,
			2+i}) // should be coords 2,3; 2,4   seq 1,2
	}

	isNotTP := rw.IsNotTeleport("1", 3, &shared.Coord{3,4}, gm)
	if !isNotTP {
		fmt.Println("Should not be a teleport")
		t.Fail()
	}

	fmt.Println("Passed")
}

func TestNoTeleportOutOfSeq(t *testing.T) {
	rw := before()
	gs := shared.InitialGameSettings{3000, 3000,
		[]shared.Coord{{1,1}, {10, 90}, {23, 99} }, 200}
	gm := geometry.CreateNewGridManager(gs)

	for i := 1; i < 3; i++ {
		rw.Add("1", uint64(i), &shared.Coord{2+i,
			2}) // should be coords 3,2; 4,2   seq 1,2
	}

	isNotTP := rw.IsNotTeleport("1", 6, &shared.Coord{4, 3}, gm)
	if !isNotTP {
		fmt.Println("Should not be a teleport")
		t.Fail()
	}

	fmt.Println("Passed")
}

func TestNoTeleportBehind(t *testing.T) {
	rw := before()
	gs := shared.InitialGameSettings{3000, 3000,
		[]shared.Coord{{1,1}, {10, 90}, {23, 99} }, 200}
	gm := geometry.CreateNewGridManager(gs)

	for i := 1; i < 3; i++ {
		rw.Add("1", uint64(i), &shared.Coord{2+i,
			2}) // should be coords 3,2; 4,2   seq 1,2
	}

	isNotTP := rw.IsNotTeleport("1", 2, &shared.Coord{3, 3}, gm)
	if !isNotTP {
		fmt.Println("Should not be a teleport")
		t.Fail()
	}

	fmt.Println("Passed")
}

func TestTeleportInSeq(t *testing.T) {
	rw := before()
	gs := shared.InitialGameSettings{3000, 3000,
		[]shared.Coord{{1,1}, {10, 90}, {23, 99} }, 200}
	gm := geometry.CreateNewGridManager(gs)

	for i := 5; i < 8; i++ {
		rw.Add("1", uint64(i), &shared.Coord{10+i,
			2}) // should be coords 15, 2; 16,2 ; 17,2   seq 5,6,7
	}

	isNotTP := rw.IsNotTeleport("1", 8, &shared.Coord{18,3}, gm)
	if isNotTP {
		fmt.Println("Should be a teleport")
		t.Fail()
	}

	fmt.Println("Passed")
}

func TestTeleportGap(t *testing.T) {
	rw := before()
	gs := shared.InitialGameSettings{3000, 3000,
		[]shared.Coord{{1,1}, {10, 90}, {23, 99} }, 200}
	gm := geometry.CreateNewGridManager(gs)

	for i := 5; i < 8; i++ {
		rw.Add("1", uint64(i), &shared.Coord{10+i,
			2}) // should be coords 15, 2; 16,2 ; 17,2   seq 5,6,7
	}

	isNotTP := rw.IsNotTeleport("1", 30, &shared.Coord{15,3}, gm)
	if isNotTP {
		fmt.Println("Should be a teleport")
		t.Fail()
	}

	fmt.Println("Passed")
}

func TestTeleportBehind(t *testing.T) {
	rw := before()
	gs := shared.InitialGameSettings{3000, 3000,
		[]shared.Coord{{1,1}, {10, 90}, {23, 99} }, 200}
	gm := geometry.CreateNewGridManager(gs)

	for i := 5; i < 8; i++ {
		rw.Add("1", uint64(i), &shared.Coord{10+i,
			2}) // should be coords 15, 2; 16,2 ; 17,2   seq 5,6,7
	}

	isNotTP := rw.IsNotTeleport("1", 6, &shared.Coord{17,3}, gm)
	if isNotTP {
		fmt.Println("Should be a teleport")
		t.Fail()
	}

	fmt.Println("Passed")
}