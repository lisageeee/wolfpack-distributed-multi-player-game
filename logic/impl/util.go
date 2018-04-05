package impl

import (
	"net"
	"log"
)

// Starts a UDP listener over the given address string, returns the address and the connection
func StartListenerUDP(ip_addr string) (*net.UDPAddr, *net.UDPConn) {
	// takes an ip address and port to listen on
	// returns the udp address and listener client
	// starts Listener
	udp_addr, _ := net.ResolveUDPAddr("udp", ip_addr)
	client, err := net.ListenUDP("udp", udp_addr)
	if err != nil {
		panic(err)
	}
	local_udp := client.LocalAddr().(*net.UDPAddr)
	local_udp.IP = GetOutboundIP()
	return local_udp, client
}

// Get the public IP of the current connection
func GetOutboundIP() net.IP {
	// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
