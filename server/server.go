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
func (foo *GServer)Register(ip string, response *shared.GameConfig) error {
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

	settings := shared.InitialGameSettings {
		WindowsX: 300,
		WindowsY: 300,
		WallCoordinates: []shared.Coord{{X: 4, Y:3}, },
	}

	initState := shared.InitialState{
		Settings: settings,
		CatchWorth: 1,
	}

	*response = shared.GameConfig{
		Connections: conns,
		Identifier: identifier,
		InitState: initState,
		GlobalServerHB: 1, // TODO; change this when working on heartbeats
		Ping: 1,
		}
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


