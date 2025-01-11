# m31config

Console application for configuration [M31-AAAX4440G Modbus module](https://www.cdebyte.com/products/M31-AAAX4440G) via
Modbus TCP.
The thing is that the standard configuration application 1) requires Windows 2) does not see
the device if it is in another network. But due to the fact that all configuration parameters are located as
Modbus registers, we can configure the device directly with Modbus commands.

```
-a string
    	current network address of the device
  -d	show debug info
  -dns string
    	set new DNS address
  -gw string
    	set default gateway address
  -ip string
    	set new device network address
  -m string
    	set subnet mask
  -p int
    	current network port of the device
  -s uint
    	device slave identifier (default 1)
```

## Getting current device configuration

```shell
./m31config -a 192.168.3.7 -p 502 -s 1
````

```
MAC : 5c:53:10:c3:67:32
DHCP: disabled
Protocol Type: TCP
Network Mode: TCP server
IP  : 192.168.3.7
Port: 502
Mask: 255.255.255.0
GW  : 192.168.3.1
DNS : 192.168.3.1

Baud rate: 9600
Parity check: NONE
```

## Set new network parameters

Current device address is `192.168.3.7`

```shell
./m31config -a 192.168.3.7 -p 502 -s 1 -ip 192.168.2.5 -m 255.255.255.0 -gw 192.168.2.1 -dns 192.168.2.1
```

```
MAC : 5c:53:10:c3:67:32
DHCP: disabled
Protocol Type: TCP
Network Mode: TCP server
IP  : 192.168.2.5
Port: 502
Mask: 255.255.255.0
GW  : 192.168.2.1
DNS : 192.168.2.1

Baud rate: 9600
Parity check: NONE
```

## Building

```shell
go build .
```