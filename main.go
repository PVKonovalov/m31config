package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/goburrow/modbus"
	"net"
	"time"
)

type Flags struct {
	Address    string
	Port       int
	SlaveId    byte
	Debug      bool
	DnsAddress string
}

var BaudRate = [...]uint{2400, 4800, 9600, 19200, 38400, 57600, 115200, 230400}
var ParityCheck = [...]string{"NONE", "ODD", "EVEN"}
var NetworkMode = [...]string{"TCP server", "TCP client", "UDP server", "UDP client"}

type Config struct {
	SerialPortRateCode              uint16
	SerialPortParityCheck           uint16
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

func (s *M31Config) CreateNewModbusTcpConnection() error {

	s.modbusTcpHandler = modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", s.flags.Address, s.flags.Port))
	s.modbusTcpHandler.Timeout = time.Duration(5) * time.Second
	s.modbusTcpHandler.IdleTimeout = time.Duration(10) * time.Second
	s.modbusTcpHandler.SlaveId = s.flags.SlaveId

	return s.modbusTcpHandler.Connect()
}

func CheckIpAddress(address string) (net.IP, error) {
	ip := net.ParseIP(address)
	if ip == nil {
		return nil, fmt.Errorf("IP address is not valid")
	}

	if ip.To4() == nil {
		return nil, fmt.Errorf("IP address is not v4")
	}
	return ip, nil
}

func (s *M31Config) SetDnsAddress(address string) error {
	ip, err := CheckIpAddress(address)

	if err != nil {
		return err
	}

	fmt.Printf("Set DNS address %+v\n", ip)
	return nil
}

func (s *M31Config) ShowConfiguration() {
	client := modbus.NewClient(s.modbusTcpHandler)

	payload, err := client.ReadHoldingRegisters(30000, 89)

	if err != nil {
		fmt.Printf("Error reading data: %v\n", err)
		return
	}

	if s.flags.Debug {
		fmt.Printf("Reading data: %v\n", payload)
	}

	reader := bytes.NewReader(payload)
	config := Config{}

	if err = binary.Read(reader, binary.BigEndian, &config); err != nil {
		fmt.Printf("failed to unmarshal: %v %v", payload, err)
		return
	}

	if s.flags.Debug {
		fmt.Printf("Config: %+v\n", config)
	}

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

	if config.NetworkMode < uint16(len(NetworkMode)) {
		fmt.Printf("Network Mode: %s\n", NetworkMode[config.NetworkMode])
	} else {
		fmt.Printf("Network Mode unknown: %d\n", config.NetworkMode)
	}

	fmt.Printf("IP  : %d.%d.%d.%d\n", config.Address[0], config.Address[1], config.Address[2], config.Address[3])
	fmt.Printf("Port: %d\n", config.Port)
	fmt.Printf("Mask: %d.%d.%d.%d\n", config.Mask[0], config.Mask[1], config.Mask[2], config.Mask[3])
	fmt.Printf("GW  : %d.%d.%d.%d\n", config.Gateway[0], config.Gateway[1], config.Gateway[2], config.Gateway[3])
	fmt.Printf("DNS : %d.%d.%d.%d\n", config.Dns[0], config.Dns[1], config.Dns[2], config.Dns[3])
	fmt.Printf("\n")
	if config.SerialPortRateCode > 0 && config.SerialPortRateCode < uint16(len(BaudRate)) {
		fmt.Printf("Baud rate: %d\n", BaudRate[config.SerialPortRateCode-1])
	} else {
		fmt.Printf("Baud rate code unknown: %d\n", config.SerialPortRateCode)
	}

	if config.SerialPortParityCheck < uint16(len(ParityCheck)) {
		fmt.Printf("Parity chek: %s\n", ParityCheck[config.SerialPortParityCheck])
	} else {
		fmt.Printf("Parity chek code unknown: %d\n", config.SerialPortParityCheck)
	}
}

func (s *M31Config) ParseFlags() {
	var slaveId uint

	flag.StringVar(&s.flags.Address, "a", "192.168.3.7", "network address")
	flag.IntVar(&s.flags.Port, "p", 502, "network port")
	flag.UintVar(&slaveId, "s", 1, "slave identifier")
	flag.BoolVar(&s.flags.Debug, "d", false, "show debug info")
	flag.StringVar(&s.flags.DnsAddress, "dns", "", "set DNS address")
	s.flags.SlaveId = byte(slaveId)
	flag.Parse()

}

func main() {

	s := NewM31Config()

	s.ParseFlags()

	if err := s.CreateNewModbusTcpConnection(); err != nil {
		fmt.Printf("Error creating modbus tcp connection: %v\n", err)
		return
	}

	s.ShowConfiguration()

	if len(s.flags.DnsAddress) != 0 {
		err := s.SetDnsAddress(s.flags.DnsAddress)
		if err != nil {
			fmt.Printf("Error setting DNS address: %v\n", err)
		}
	}
}
