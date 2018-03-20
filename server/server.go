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
)

type GServer int

type Player struct {
	Address net.Addr
	RecentHB int64
}

type AllPlayers struct {
	sync.RWMutex
	all map[string]*Player
}

var (
	heartBeat = uint32(500)
	ping = uint32(3)
	allPlayers = AllPlayers{all: make(map[string]*Player)}
)

type PlayerInfo struct {
	Address net.Addr
	Key ecdsa.PublicKey
}

func main() {
	gob.Register(&net.UDPAddr{})

	gserver := new(GServer)

	server := rpc.NewServer()
	server.Register(gserver)

	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}

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
	allPlayers.Lock()
	defer allPlayers.Unlock()

	pubKeyStr := "hah" // TODO: replace with key-generators pubKeyToString
	if player, exists := allPlayers.all[pubKeyStr]; exists {
		fmt.Println("DEBUG - Key Already Registered Error")
		return wolferrors.KeyAlreadyRegisteredError(player.Address.String())
	}

	for _, player := range allPlayers.all {
		if player.Address.Network() == p.Address.Network() && player.Address.String() == p.Address.String() {
			fmt.Println("DEBUG - Address Already Registered Error")
			return wolferrors.AddressAlreadyRegisteredError(p.Address.String())
		}
	}

	// once all checks are made to ensure that this connecting player has not already been registered,
	// add this player to allPlayers struct
	allPlayers.all[pubKeyStr] = &Player {
		p.Address,
		time.Now().UnixNano(),
	}

	go monitor(pubKeyStr, time.Duration(heartBeat)*time.Millisecond)

	settings := shared.InitialGameSettings {
		WindowsX: 300,
		WindowsY: 300,
		WallCoordinates: []shared.Coord{{X: 4, Y:3}, },
	}

	initState := shared.InitialState {
		Settings: settings,
		CatchWorth: 1,
	}

	*response = shared.GameConfig {
		InitState: initState,
		GlobalServerHB: heartBeat,
		Ping: ping,
		}

	return nil
}

func (foo *GServer) GetNodes(key ecdsa.PublicKey, addrSet *[]net.Addr) error {
	allPlayers.RLock()
	defer allPlayers.RUnlock()

	pubKeyStr := "hah" // TODO: replace with key-generators pubKeyToString

	if _, ok := allPlayers.all[pubKeyStr]; !ok {
		fmt.Println("DEBUG - Unknown Key Error")
		return wolferrors.UnknownKeyError(pubKeyStr)
	}

	playerAddresses := make([]net.Addr, 0, len(allPlayers.all) - 1)

	for k, player := range allPlayers.all {
		if k == pubKeyStr {
			continue
		}
		playerAddresses = append(playerAddresses, player.Address)
	}

	n := len(playerAddresses)
	*addrSet = playerAddresses[:n]

	return nil
}

func (foo *GServer) Heartbeat(key ecdsa.PublicKey, _ignored *bool) error {
	allPlayers.Lock()
	defer allPlayers.Unlock()

	pubKeyStr := "hah" // TODO: replace with key-generators pubKeyToString

	if _, ok := allPlayers.all[pubKeyStr]; !ok {
		fmt.Println("DEBUG - Unknown Key Error")
		return wolferrors.UnknownKeyError(pubKeyStr)
	}

	allPlayers.all[pubKeyStr].RecentHB = time.Now().UnixNano()

	return nil
}
