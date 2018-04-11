package test
import (
	l "../logic/impl"
	"testing"
	"../shared"
	"fmt"
	"time"
)

func before()l.RunningWindow {
	return l.RunningWindow{Map:make(map[string][3]l.MoveSeq)}
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