package main

import (
	"net/rpc"
	"net"
	"../shared"
	"sync"
	"crypto/ecdsa"
	"../wolferrors"
	"time"
	"encoding/gob"
	"fmt"
	"crypto/elliptic"
	"strconv"
	"os"
	keys "../key-helpers"
)

// Usage go run server.go (runs on port 8081) or go run server.go [portnumber]

type GServer struct {
	SelectConfig string
}

type Player struct {
	Address net.Addr
	RecentHB int64
	Identifier int
}

type AllPlayers struct {
	sync.RWMutex
	all map[string]*Player
}

var (
	heartBeat = uint32(5000)
	ping = uint32(3)
	id = 0
	allPlayers = AllPlayers{all: make(map[string]*Player)}
)

type PlayerInfo struct {
	Address net.Addr
	PubKey ecdsa.PublicKey
}

func main() {
	portString := ":8081"
	configString := "0"
	args := os.Args
	if len(args) > 2 {
		portString = ":" + args[1]
		configString = args[2]
	} else if len(args) > 1 {
		portString = ":" + args[1]
	}
	gob.Register(&net.UDPAddr{})
	gob.Register(&elliptic.CurveParams{})
	gob.Register(&PlayerInfo{})

	gserver := new(GServer)
	gserver.SelectConfig = configString

	server := rpc.NewServer()
	server.Register(gserver)

	l, err := net.Listen("tcp", portString)
	if err != nil {
		fmt.Printf("Server: error listening for incoming connections on port [%s]. Ensure there is not another" +
			" server already running", portString)
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, _ := l.Accept()
		go server.ServeConn(conn)
	}
}

func monitor(pubKeyStr string, heartBeatInterval time.Duration) {
	for {
		allPlayers.Lock()
		if time.Now().UnixNano() - allPlayers.all[pubKeyStr].RecentHB > int64(heartBeatInterval) {
			fmt.Printf("Disconnected and deleted: %s\n", allPlayers.all[pubKeyStr].Address.String())
			delete(allPlayers.all, pubKeyStr)
			allPlayers.Unlock()
			return
		}
		allPlayers.Unlock()
		time.Sleep(heartBeatInterval)
	}
}

func (foo *GServer) Register(p PlayerInfo, response *shared.GameConfig) error {
	id++
	allPlayers.Lock()
	defer allPlayers.Unlock()

	pubKeyStr := keys.PubKeyToString(p.PubKey)

	// TODO: This needs to be fixed
	//if player, exists := allPlayers.all[pubKeyStr]; exists {
	//	fmt.Printf("DEBUG - Key Already Registered Error [%s]\n",
	//		player.Address.String())
	//	return wolferrors.KeyAlreadyRegisteredError(player.Address.String())
	//}

	for _, player := range allPlayers.all {
		if player.Address.Network() == p.Address.Network() && player.Address.String() == p.Address.String() {
			fmt.Printf("DEBUG - Address Already Registered Error [%s], [%s]\n",
				player.Address.Network(), player.Address.String())
			return wolferrors.AddressAlreadyRegisteredError(p.Address.String())
		}
	}

	// once all checks are made to ensure that this connecting player has not already been registered,
	// add this player to allPlayers struct
	allPlayers.all[pubKeyStr] = &Player {
		Address: p.Address,
		RecentHB: time.Now().UnixNano(),
		Identifier: id,
	}

	fmt.Printf("DEBUG - [%s] Connected\n", p.Address.String())

	go monitor(pubKeyStr, time.Duration(heartBeat)*time.Millisecond)

	settings := getSettingsByConfigString(foo.SelectConfig)
	*response = settings

	return nil
}

func (foo *GServer) RegisterPrey(p PlayerInfo, response *shared.GameConfig) error {
	allPlayers.Lock()
	defer allPlayers.Unlock()

	pubKeyStr := keys.PubKeyToString(p.PubKey)

	// TODO: This needs to be fixed
	//if player, exists := allPlayers.all[pubKeyStr]; exists {
	//	fmt.Printf("DEBUG - Key Already Registered Error [%s]\n",
	//		player.Address.String())
	//	return wolferrors.KeyAlreadyRegisteredError(player.Address.String())
	//}

	for _, player := range allPlayers.all {
		if player.Address.Network() == p.Address.Network() && player.Address.String() == p.Address.String() {
			fmt.Printf("DEBUG - Address Already Registered Error [%s], [%s]\n",
				player.Address.Network(), player.Address.String())
			return wolferrors.AddressAlreadyRegisteredError(p.Address.String())
		}
	}

	// once all checks are made to ensure that this connecting player has not already been registered,
	// add this player to allPlayers struct
	allPlayers.all[pubKeyStr] = &Player {
		Address: p.Address,
		RecentHB: time.Now().UnixNano(),
		Identifier: -1,
	}

	fmt.Printf("DEBUG - [%s] Connected\n", p.Address.String())

	go monitor(pubKeyStr, time.Duration(heartBeat)*time.Millisecond)

	settings := getSettingsByConfigString(foo.SelectConfig)
	*response = settings

	return nil
}

func (foo *GServer) GetNodes(key ecdsa.PublicKey, addrSet * map[string]shared.NodeRegistrationInfo) error {
	allPlayers.RLock()
	defer allPlayers.RUnlock()

	pubKeyStr := keys.PubKeyToString(key)

	if _, ok := allPlayers.all[pubKeyStr]; !ok {
		fmt.Println("DEBUG - Unknown Key Error")
		return wolferrors.UnknownKeyError(pubKeyStr)
	}

	playerAddresses := make(map[string]shared.NodeRegistrationInfo)

	for k, player := range allPlayers.all {
		if k == pubKeyStr {
			continue
		}
		idString := strconv.Itoa(player.Identifier)
		playerAddresses[idString] = shared.NodeRegistrationInfo{Id: idString, Addr: player.Address, PubKey: k}
	}

	*addrSet = playerAddresses

	return nil
}

func (foo *GServer) Heartbeat(key ecdsa.PublicKey, _ignored *bool) error {
	allPlayers.Lock()
	defer allPlayers.Unlock()

	pubKeyStr := keys.PubKeyToString(key)

	if _, ok := allPlayers.all[pubKeyStr]; !ok {
		fmt.Println("DEBUG - Unknown Key Error")
		return wolferrors.UnknownKeyError(pubKeyStr)
	}

	allPlayers.all[pubKeyStr].RecentHB = time.Now().UnixNano()

	return nil
}

func getSettingsByConfigString(configString string) (shared.GameConfig) {
	var response shared.GameConfig
	switch configString {
	case "1":
		settings := shared.InitialGameSettings {
			WindowsX: 600,
			WindowsY: 600,
			WallCoordinates: []shared.Coord{
				// Left side
				{X: 0, Y:0}, {X: 0, Y:1}, {X: 0, Y:2}, {X: 0, Y:3}, {X: 0, Y:4}, {X: 0, Y:5},{X: 0, Y:6}, {X: 0, Y:7},
				{X: 0, Y:8}, {X: 0, Y:9}, {X: 0, Y:10}, {X: 0, Y:11},{X: 0, Y:12}, {X: 0, Y:13}, {X: 0, Y:14}, {X: 0, Y:15},
				{X: 0, Y:16}, {X: 0, Y:17}, {X: 0, Y:18}, {X: 0, Y:19},
				// Right side
				{X:19, Y:0}, {X:19, Y:1}, {X:19, Y:2}, {X:19, Y:3}, {X:19, Y:4}, {X:19, Y:5},{X:19, Y:6}, {X:19, Y:7},
				{X:19, Y:8}, {X:19, Y:9}, {X:19, Y:10}, {X:19, Y:11},{X:19, Y:12}, {X:19, Y:13}, {X:19, Y:14}, {X:19, Y:15},
				{X:19, Y:16}, {X:19, Y:17}, {X:19, Y:18}, {X:19, Y:19},
				//Bottom
				{X: 1, Y:0}, {X: 2, Y:0}, {X: 3, Y:0}, {X: 4, Y:0}, {X: 5, Y:0}, {X: 6, Y:0},{X: 7, Y:0}, {X: 8, Y:0},
				{X: 9, Y:0}, {X: 10, Y:0}, {X: 11, Y:0}, {X: 12, Y:0},{X: 13, Y:0}, {X:14, Y:0}, {X:15, Y:0}, {X: 16, Y:0},
				{X: 17, Y:0}, {X: 18, Y:0}, {X: 19, Y:0},
				// Top
				{X: 1, Y:19}, {X: 2, Y:19}, {X: 3, Y:19}, {X: 4, Y:19}, {X: 5, Y:19}, {X: 6, Y:19},{X: 7, Y:19}, {X: 8, Y:19},
				{X: 9, Y:19}, {X: 10, Y:19}, {X: 11, Y:19}, {X: 12, Y:19},{X: 13, Y:19}, {X:14, Y:19}, {X:15, Y:19}, {X: 16, Y:19},
				{X: 17, Y:19}, {X: 18, Y:19}, {X: 19, Y:19},
				// Draw inside from top to bottom, then left to right
				{X: 1, Y:16}, {X: 2, Y:16},
				{X: 2, Y:11}, {X: 3, Y:11}, {X: 4, Y:11},
				{X: 2, Y:12}, {X: 3, Y:12}, {X: 4, Y:12},
				{X: 5, Y:7}, {X: 5, Y:8}, {X: 5, Y:9}, {X: 4, Y:8}, {X: 6, Y:7},
				{X: 4, Y:3}, {X: 5, Y:3}, {X: 4, Y:4}, {X: 5, Y:4},{X: 3, Y:3},{X: 4, Y:2},
				{X: 7, Y:17},{X: 8, Y:17}, {X: 8, Y:16}, {X: 9, Y:16},
				{X: 8, Y:13}, {X: 9, Y:13},
				{X: 10, Y:10}, {X: 11, Y:10},{X: 12, Y:10},{X: 11, Y:9},
				{X: 12, Y:5}, {X: 11, Y:4},{X: 12, Y:4},{X: 13, Y:4},{X: 12, Y:3},{X: 13, Y:3},{X: 14, Y:3},
				{X: 12, Y:1},
				{X: 13, Y:17},{X: 14, Y:16},
				{X: 17, Y:14},{X: 16, Y:13},{X: 15, Y:12},
				{X: 15, Y:9},{X: 15, Y:8},{X: 14, Y:8},{X: 14, Y:7},{X: 13, Y:7},
				{X: 18, Y:5},{X: 18, Y:4},{X: 18, Y:3},
				},
			ScoreboardWidth: 200,
		}

		initState := shared.InitialState {
			Settings: settings,
			CatchWorth: 1,
		}

		response = shared.GameConfig {
			InitState: 	initState,
			Identifier: id,
			GlobalServerHB: heartBeat,
			Ping: 		ping,
		}
	default:
		settings := shared.InitialGameSettings {
			WindowsX: 300,
			WindowsY: 300,
			WallCoordinates: []shared.Coord{{X: 4, Y:3}, {X: 9, Y:9}},
			ScoreboardWidth: 200,
		}

		initState := shared.InitialState {
			Settings: settings,
			CatchWorth: 1,
		}

		response = shared.GameConfig {
			InitState: 	initState,
			Identifier: id,
			GlobalServerHB: heartBeat,
			Ping: 		ping,
		}
	}

	return response
}
func pubKeyToString(key ecdsa.PublicKey) string {
	return string(elliptic.Marshal(key.Curve, key.X, key.Y))
}