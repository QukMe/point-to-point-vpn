package main

import (
	"flag"
	"log"
	"net"
	"sync/atomic"

	"p2p-vpn/internal/crypto"
	"p2p-vpn/internal/tun"
	"p2p-vpn/internal/tunnel"
)

func main() {
	port := flag.Int("port", 2000, "udp port")
	key := flag.String("key", "", "encryption key (32 bytes hex)")
	flag.Parse()

	aead := crypto.NewAEAD(*key)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: *port})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	device := tun.Create()
	defer device.Close()

	var remote atomic.Pointer[net.UDPAddr]
	tunnel.Run(device, conn, aead, &remote)
}
