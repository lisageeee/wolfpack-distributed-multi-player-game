package impl

import "net"

func StartListenerUDP(ip_addr string) (*net.UDPAddr, *net.UDPConn) {
	// takes an ip address and port to listen on
	// returns the udp address and listener client
	// starts Listener
	udp_addr, _ := net.ResolveUDPAddr("udp", ip_addr)
	client, err := net.ListenUDP("udp", udp_addr)
	if err != nil {
		panic(err)
	}
	return udp_addr, client
}

func StartSenderUDP(ip_addr string) (*net.UDPConn) {
	node_udp, _ := net.ResolveUDPAddr("udp", ip_addr)
	node_client, err := net.DialUDP("udp", nil, node_udp)
	if err != nil {
		panic(err)
	}
	return node_client
}