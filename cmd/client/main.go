package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"p2p-vpn/internal/crypto"
	"p2p-vpn/internal/tun"
	"p2p-vpn/internal/tunnel"
)

func main() {
	server := flag.String("server", "", "server public ip")
	port := flag.Int("port", 2000, "server udp port")
	key := flag.String("key", "", "encryption key (32 bytes hex)")
	flag.Parse()

	aead := crypto.NewAEAD(*key)

	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *server, *port))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	device := tun.Create()
	defer device.Close()

	var remote atomic.Pointer[net.UDPAddr]
	remote.Store(serverAddr)
	tunnel.Run(device, conn, aead, &remote)
}
