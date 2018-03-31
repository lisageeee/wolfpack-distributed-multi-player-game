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

func (foo *GServer) GetNodes(key ecdsa.PublicKey, addrSet * map[string]net.Addr) error {
	allPlayers.RLock()
	defer allPlayers.RUnlock()

	pubKeyStr := keys.PubKeyToString(key)

	if _, ok := allPlayers.all[pubKeyStr]; !ok {
		fmt.Println("DEBUG - Unknown Key Error")
		return wolferrors.UnknownKeyError(pubKeyStr)
	}

	playerAddresses := make(map[string]net.Addr)

	for k, player := range allPlayers.all {
		if k == pubKeyStr {
			continue
		}
		playerAddresses[strconv.Itoa(player.Identifier)] = player.Address
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
			WallCoordinates: []shared.Coord{{X: 4, Y:3}, {X: 9, Y:9}, {X: 4, Y:4}, {X: 4, Y:5}},
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