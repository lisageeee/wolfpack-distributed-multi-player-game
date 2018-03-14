package main

import (
	"net/rpc"
	"net"
	"fmt"
	"../shared"
)

type GServer int

var conns []string

func main() {
	gserver := new(GServer)
	s := rpc.NewServer()
	s.Register(gserver)
	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	for {
		conn, _ := l.Accept()
		go s.ServeConn(conn)
	}
}
func (foo *GServer)Register(ip string, response *shared.RegistrationDetails) error {
	fmt.Println("Got connection from: ", ip)

	var identifier int
	if !hasIP(conns, ip) {
		fmt.Println("adding connection")
		conns = append(conns, ip)
		identifier = len(conns)
	} else {
		for i, conn := range conns {
			if conn == ip {
				identifier = i
			}
		}
	}

	// TODO: let the user specify these
	settings := shared.EnvironmentSettings{
		WinMaxX: 300,
		WinMaxY: 300,
		WallCoords: []shared.Coord{{X: 1, Y:2}, {X: 1, Y:3}},
	}

	initState := shared.InitialState{
		Settings: settings,
		CatchWorth: 1,
	}

	*response = shared.RegistrationDetails{Connections: conns, Identifier: identifier, InitState: initState}
	return nil
}

func hasIP(conns []string, toMatch string ) bool {
	for _, val := range conns{
		if val == toMatch{
			return true
		}
	}
	return false
}


