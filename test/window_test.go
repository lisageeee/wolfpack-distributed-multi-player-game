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

	rw.Add("id1", uint64(1), &shared.Coord{10, 6})
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
	fmt.Println("Add Five Test Multiple Ids")

	for i := 1; i < 6; i++ {
		go rw.Add("id1", uint64(i), &shared.Coord{i+4,
			i-2})
	}
	for i := 1; i < 6; i++ {
		go rw.Add("id2", uint64(i), &shared.Coord{i+4,
			i-2})
	}
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
