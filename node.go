package main

import (
	"fmt"
	"net"
	"os"
	"net/rpc"
	"strconv"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image"

	_ "image/png"
	_ "image/jpeg"
)


func main() {
	fmt.Println("hello world")

	// Listener IP address
	var ip_addr string
	// Can start with an IP as param
	if (len(os.Args)>1){
		ip_addr = os.Args[1]
	}else{
		ip_addr = "127.0.0.1:0"
	}
	_, client := startListener(ip_addr)
	defer client.Close()

	go RunListener(client)

	otherNodes := serverRegister(client.LocalAddr().String())
	udpAddr := client.LocalAddr().(*net.UDPAddr)
	floodNodes(otherNodes, udpAddr)

	pixelgl.Run(run)
	select {}
}
func run() {
	// all of our code will be fired up from here
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.Clear(colornames.Skyblue)
	pic, err := loadPicture("bunny.jpeg")
	if err != nil {
		panic(err)
	}
	sprite := pixel.NewSprite(pic, pixel.R(20,20,50,50))
	win.Clear(colornames.Skyblue)
	center:= win.Bounds().Center()
	sprite.Draw(win, pixel.IM.Moved(center))

	for !win.Closed() {

		if win.Pressed(pixelgl.KeyLeft){
			win.Clear(colornames.Skyblue)
			mat := pixel.IM
			center.X = center.X-1
			mat = mat.Moved(center)
			sprite.Draw(win, mat)
		}
		if win.Pressed(pixelgl.KeyRight){

			win.Clear(colornames.Skyblue)
			mat := pixel.IM
			center.X = center.X+1
			mat = mat.Moved(center)
			sprite.Draw(win, mat)
		}
		if win.Pressed(pixelgl.KeyUp){
			win.Clear(colornames.Skyblue)
			mat := pixel.IM
			center.Y = center.Y+1
			mat = mat.Moved(center)
			sprite.Draw(win, mat)
		}
		if win.Pressed(pixelgl.KeyDown){
			win.Clear(colornames.Skyblue)
			mat := pixel.IM
			center.Y = center.Y-1
			mat = mat.Moved(center)
			sprite.Draw(win, mat)
		}
		win.Update()

	}

}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func floodNodes(otherNodes []string, udp_addr *net.UDPAddr) {
	localIP, _ := net.ResolveUDPAddr("udp", udp_generic)
	for _, ip := range otherNodes {
		node_udp, _ := net.ResolveUDPAddr("udp", ip)
		// Connect to other node
		node_client, err := net.DialUDP("udp", localIP, node_udp)
		if err != nil {
			panic(err)
		}
		// Exchange messages with other node
		myListener := udp_addr.IP.String() + ":" +  strconv.Itoa(udp_addr.Port)
		node_client.Write([]byte(myListener))
	}
}

func serverRegister(localIP string) []string {
	// Connect to server with RPC, port is always :8081
	serverConn, err := rpc.Dial("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	var response []string
	// Get IP from server
	err = serverConn.Call("GServer.Register", localIP, &response)
	if err != nil {
		panic(err)
	}
	if len(response) > 0 {
		for ind, val := range response {
			fmt.Println(strconv.Itoa(ind) + ": " + val)
		}
	}
	return response
}

func startListener(ip_addr string) (*net.UDPAddr, *net.UDPConn) {
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

const udp_generic = "127.0.0.1:0"
var clients []*net.Conn
func RunListener(client *net.UDPConn) {
	// takes a listener client
	// runs the listener in a infinite loop

	client.SetReadBuffer(1048576)

	i := 0
	for {
		i++
		buf := make([]byte, 1024)
		rlen, addr, err := client.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(buf[0:rlen]))
		fmt.Println(addr)
		fmt.Println(i)
		if string(buf[0:rlen]) != "connected" {
			remote_client, err := net.Dial("udp", string(buf[0:rlen]))
			if err != nil {
				panic(err)
			}
			remote_client.Write([]byte("connected"))

			clients = append(clients, &remote_client)
		}
	}
}

