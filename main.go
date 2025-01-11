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
	Address        string
	Port           int
	SlaveId        byte
	Debug          bool
	DnsAddress     string
	DeviceAddress  string
	GatewayAddress string
	SubnetMask     string
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
	SubnetMask                      [4]byte
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

	if ipv4 := ip.To4(); ipv4 == nil {
		return nil, fmt.Errorf("IP address is not v4")
	} else {
		return ipv4, nil
	}
}

func (s *M31Config) SetDnsAddress(address string) error {
	return s.setIpAddress(address, 30013)
}

func (s *M31Config) SetGatewayAddress(address string) error {
	return s.setIpAddress(address, 30011)
}

func (s *M31Config) SetDeviceAddress(address string) error {
	return s.setIpAddress(address, 30007)
}

func (s *M31Config) SetSubnetMask(address string) error {
	return s.setIpAddress(address, 30009)
}

func (s *M31Config) setIpAddress(address string, firstRegister uint16) error {
	ip, err := CheckIpAddress(address)

	if err != nil {
		return err
	}

	reg1 := uint16(ip[0])<<8 | uint16(ip[1])
	reg2 := uint16(ip[2])<<8 | uint16(ip[3])

	client := modbus.NewClient(s.modbusTcpHandler)

	_, err = client.WriteSingleRegister(firstRegister, reg1)
	if err != nil {
		return err
	}

	_, err = client.WriteSingleRegister(firstRegister+1, reg2)

	return err
}

func (s *M31Config) ShowConfiguration() {
	newValues := ""

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

	if len(s.flags.DeviceAddress) != 0 {
		newValues = fmt.Sprintf(" → %s", s.flags.DeviceAddress)
	} else {
		newValues = ""
	}
	fmt.Printf("IP  : %d.%d.%d.%d%s\n", config.Address[0], config.Address[1], config.Address[2], config.Address[3], newValues)
	fmt.Printf("Port: %d\n", config.Port)
	if len(s.flags.SubnetMask) != 0 {
		newValues = fmt.Sprintf(" → %s", s.flags.SubnetMask)
	} else {
		newValues = ""
	}
	fmt.Printf("Mask: %d.%d.%d.%d%s\n", config.SubnetMask[0], config.SubnetMask[1], config.SubnetMask[2], config.SubnetMask[3], newValues)

	if len(s.flags.GatewayAddress) != 0 {
		newValues = fmt.Sprintf(" → %s", s.flags.GatewayAddress)
	} else {
		newValues = ""
	}
	fmt.Printf("GW  : %d.%d.%d.%d%s\n", config.Gateway[0], config.Gateway[1], config.Gateway[2], config.Gateway[3], newValues)

	if len(s.flags.DnsAddress) != 0 {
		newValues = fmt.Sprintf(" → %s", s.flags.DnsAddress)
	} else {
		newValues = ""
	}
	fmt.Printf("DNS : %d.%d.%d.%d%s\n", config.Dns[0], config.Dns[1], config.Dns[2], config.Dns[3], newValues)
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

	flag.StringVar(&s.flags.Address, "a", "", "current network address of the device")
	flag.IntVar(&s.flags.Port, "p", 0, "current network port of the device")
	flag.UintVar(&slaveId, "s", 1, "device slave identifier")
	flag.BoolVar(&s.flags.Debug, "d", false, "show debug info")
	flag.StringVar(&s.flags.DnsAddress, "dns", "", "set new DNS address")
	flag.StringVar(&s.flags.DeviceAddress, "ip", "", "set new device network address")
	flag.StringVar(&s.flags.GatewayAddress, "gw", "", "set default gateway address")
	flag.StringVar(&s.flags.SubnetMask, "m", "", "set subnet mask")
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

	if len(s.flags.GatewayAddress) != 0 {
		err := s.SetGatewayAddress(s.flags.GatewayAddress)
		if err != nil {
			fmt.Printf("Error setting default gateway address: %v\n", err)
		}
	}

	if len(s.flags.DeviceAddress) != 0 {
		err := s.SetDeviceAddress(s.flags.DeviceAddress)
		if err != nil {
			fmt.Printf("Error setting device address: %v\n", err)
		}
	}
	if len(s.flags.SubnetMask) != 0 {
		err := s.SetSubnetMask(s.flags.SubnetMask)
		if err != nil {
			fmt.Printf("Error setting subnet mask: %v\n", err)
		}
	}

	s.ShowConfiguration()
}
