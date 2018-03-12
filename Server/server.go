package main

import (
	"net/rpc"
	"net"
	"fmt"
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
func (foo *GServer)Register(ip string, response *[]string) error {
	fmt.Println("Got connection from: ", ip)
	*response = conns
	if !hasIP(conns, ip) {
		fmt.Println("adding connection")
		conns = append(conns, ip)
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


