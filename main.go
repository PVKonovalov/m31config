package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	//"m31config/pkg/modbus"
	"github.com/goburrow/modbus"
	"time"
)

type Flags struct {
	Address string
	Port    int
	SlaveId byte
}

type Config struct {
	SerialPortRateCode              uint16
	SerialPortChecking              uint16
	NetworkMode                     uint16
	Dhcp                            uint16
	MacAddress                      [6]byte
	Address                         [4]byte
	Mask                            [4]byte
	Gateway                         [4]byte
	Dns                             [4]byte
	Free                            [4]byte
	Port                            uint16
	DestinationDomainName           [128]byte
	DestinationPort                 uint16
	ProtocolType                    uint16
	AddressNegotiationWriteRegister uint16
	AddressNegotiationRegister      uint16
	NegotiationStatus               uint16
	ExceptionCode                   uint16
}

type M31Config struct {
	flags            Flags
	modbusTcpHandler *modbus.TCPClientHandler
}

func NewM31Config() *M31Config {
	return &M31Config{}
}

func (c *M31Config) CreateNewModbusTcpConnection() error {

	c.modbusTcpHandler = modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", c.flags.Address, c.flags.Port))
	c.modbusTcpHandler.Timeout = time.Duration(5) * time.Second
	c.modbusTcpHandler.IdleTimeout = time.Duration(10) * time.Second
	c.modbusTcpHandler.SlaveId = c.flags.SlaveId

	return c.modbusTcpHandler.Connect()
}

func (c *M31Config) ShowConfiguration() {
	client := modbus.NewClient(c.modbusTcpHandler)

	payload, err := client.ReadHoldingRegisters(30000, 89)

	if err != nil {
		fmt.Printf("Error reading data: %v\n", err)
		return
	}

	fmt.Printf("Reading data: %v\n", payload)

	reader := bytes.NewReader(payload)
	config := Config{}

	if err = binary.Read(reader, binary.BigEndian, &config); err != nil {
		fmt.Printf("failed to unmarshal: %v %v", payload, err)
		return
	}

	fmt.Printf("Config: %+v\n", config)
	fmt.Printf("MAC : %x:%x:%x:%x:%x:%x\n", config.MacAddress[0], config.MacAddress[1], config.MacAddress[2], config.MacAddress[3], config.MacAddress[4], config.MacAddress[5])
	if config.Dhcp == 1 {
		fmt.Printf("DHCP: enabled\n")
	} else {
		fmt.Printf("DHCP: disabled\n")
	}
	if config.ProtocolType == 1 {
		fmt.Printf("Protocol Type: TCP\n")
	} else {
		fmt.Printf("Protocol Type: RTU\n")
	}

	switch config.NetworkMode {
	case 0:
		fmt.Printf("Mode: TCP server\n")
	case 1:
		fmt.Printf("Mode: TCP client\n")
	case 2:
		fmt.Printf("Mode: UDP server\n")
	case 3:
		fmt.Printf("Mode: UDP client\n")
	default:
		fmt.Printf("Mode: Unknown (%d)\n", config.NetworkMode)
	}
	fmt.Printf("IP  : %d.%d.%d.%d\n", config.Address[0], config.Address[1], config.Address[2], config.Address[3])
	fmt.Printf("Port: %d\n", config.Port)
	fmt.Printf("Mask: %d.%d.%d.%d\n", config.Mask[0], config.Mask[1], config.Mask[2], config.Mask[3])
	fmt.Printf("GW  : %d.%d.%d.%d\n", config.Gateway[0], config.Gateway[1], config.Gateway[2], config.Gateway[3])
	fmt.Printf("DNS : %d.%d.%d.%d\n", config.Dns[0], config.Dns[1], config.Dns[2], config.Dns[3])

}

func main() {

	s := NewM31Config()

	flag.StringVar(&s.flags.Address, "a", "192.168.3.7", "network address")
	flag.IntVar(&s.flags.Port, "p", 502, "network port")
	var slaveId uint
	flag.UintVar(&slaveId, "s", 1, "slave identifier")
	s.flags.SlaveId = byte(slaveId)

	if err := s.CreateNewModbusTcpConnection(); err != nil {
		fmt.Printf("Error creating modbus tcp connection: %v", err)
		return
	}

	s.ShowConfiguration()
}
