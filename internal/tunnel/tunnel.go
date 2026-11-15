package tunnel

import (
	"crypto/cipher"
	"log"
	"net"
	"sync/atomic"

	wgtun "golang.zx2c4.com/wireguard/tun"

	"p2p-vpn/internal/crypto"
	"p2p-vpn/internal/tun"
)

func Run(device wgtun.Device, conn *net.UDPConn, aead cipher.AEAD, remote *atomic.Pointer[net.UDPAddr]) {
	go tunToUDP(device, conn, aead, remote)
	udpToTun(device, conn, aead, remote)
}

func tunToUDP(device wgtun.Device, conn *net.UDPConn, aead cipher.AEAD, remote *atomic.Pointer[net.UDPAddr]) {
	buffers := tun.NewBuffers()
	sizes := make([]int, tun.BatchSize)
	for {
		packets, err := tun.ReadPackets(device, buffers, sizes)
		if err != nil {
			log.Println(err)
			continue
		}
		addr := remote.Load()
		if addr == nil {
			continue
		}
		for _, packet := range packets {
			_, err := conn.WriteToUDP(crypto.Encrypt(aead, packet), addr)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func udpToTun(device wgtun.Device, conn *net.UDPConn, aead cipher.AEAD, remote *atomic.Pointer[net.UDPAddr]) {
	buffer := make([]byte, tun.BufferSize)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println(err)
			continue
		}
		packet, err := crypto.Decrypt(aead, buffer[:n])
		if err != nil {
			log.Println(err)
			continue
		}
		remote.Store(addr)
		err = tun.WritePackets(device, [][]byte{packet})
		if err != nil {
			log.Println(err)
		}
	}
}
