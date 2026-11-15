package tun

import (
	"log"
	"runtime"

	wgtun "golang.zx2c4.com/wireguard/tun"
)

const (
	DefaultMTU   = 1420
	BufferSize   = 2048
	BatchSize    = 32
	BufferOffset = 32
	DarwinName   = "utun4"
	LinuxName    = "tun0"
)

func interfaceName() string {
	if runtime.GOOS == "darwin" {
		return DarwinName
	}
	return LinuxName
}

func Create() wgtun.Device {
	device, err := wgtun.CreateTUN(interfaceName(), DefaultMTU)
	if err != nil {
		log.Fatal(err)
	}
	return device
}

func NewBuffers() [][]byte {
	buffers := make([][]byte, BatchSize)
	for i := range buffers {
		buffers[i] = make([]byte, BufferSize)
	}
	return buffers
}

func ReadPackets(device wgtun.Device, buffers [][]byte, sizes []int) ([][]byte, error) {
	n, err := device.Read(buffers, sizes, BufferOffset)
	if err != nil {
		return nil, err
	}
	packets := make([][]byte, n)
	for i := 0; i < n; i++ {
		packets[i] = buffers[i][BufferOffset : BufferOffset+sizes[i]]
	}
	return packets, nil
}

func WritePackets(device wgtun.Device, packets [][]byte) error {
	buffers := make([][]byte, len(packets))
	for i, packet := range packets {
		buffers[i] = make([]byte, BufferOffset+len(packet))
		copy(buffers[i][BufferOffset:], packet)
	}
	_, err := device.Write(buffers, BufferOffset)
	return err
}
