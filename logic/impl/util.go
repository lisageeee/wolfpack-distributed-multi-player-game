package impl

import "net"

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
	return client.LocalAddr().(*net.UDPAddr), client
}
