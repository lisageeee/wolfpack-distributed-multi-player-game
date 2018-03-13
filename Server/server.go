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
	if !hasIP(conns, ip) {
		fmt.Println("adding connection")
		conns = append(conns, ip)
	}
	*response = shared.RegistrationDetails{Connections: conns, Identifier: len(conns)}
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


