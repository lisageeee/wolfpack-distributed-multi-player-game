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
	"../key-helpers"
	"crypto/elliptic"
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
	id = 1
	allPlayers = AllPlayers{all: make(map[string]*Player)}
)

type PlayerInfo struct {
	Address net.Addr
	PubKey ecdsa.PublicKey
}

func main() {
	gob.Register(&net.UDPAddr{})
	gob.Register(&elliptic.CurveParams{})
	gob.Register(&PlayerInfo{})

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

	fmt.Printf("DEBUG - allPlayers [%v]\n", allPlayers.all)

	pubKeyStr := key_helpers.EncodePubKey(&p.PubKey)

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
		p.Address,
		time.Now().UnixNano(),
	}

	fmt.Printf("DEBUG - [%s] Connected\n", p.Address.String())

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
		InitState: 	initState,
		Identifier: id,
		GlobalServerHB: heartBeat,
		Ping: 		ping,
	}

	id++

	return nil
}

func (foo *GServer) GetNodes(key ecdsa.PublicKey, addrSet *[]net.Addr) error {
	allPlayers.RLock()
	defer allPlayers.RUnlock()

	pubKeyStr := key_helpers.EncodePubKey(&key)

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

	pubKeyStr := key_helpers.EncodePubKey(&key)

	if _, ok := allPlayers.all[pubKeyStr]; !ok {
		fmt.Println("DEBUG - Unknown Key Error")
		return wolferrors.UnknownKeyError(pubKeyStr)
	}

	allPlayers.all[pubKeyStr].RecentHB = time.Now().UnixNano()

	return nil
}