package main

import (
	"fmt"
	"os"
	"net"
	"strconv"
)

func main() {
	var nodePort int
	if len(os.Args) > 1 {
		port, error := strconv.Atoi(os.Args[1])
		if error != nil {
			port = 20000
		}
		nodePort = port
	} else {
		nodePort = 20000
	}
	network := Listen("0.0.0.0", nodePort)

	fmt.Println("Node initalized on port ", nodePort , "!");

	go func() {
		buffer := make([]byte, 256)
		for {
			length, address, _ := network.connection.ReadFromUDP(buffer)
			fmt.Println("Received \"" + string(buffer[0:length]) + "\" from " + address.IP.String() + ":" + strconv.Itoa(address.Port))
		}
	}()

	for {
		var input string

		fmt.Println("Enter IP to ping: ")
		fmt.Scanln(&input)
		ip := net.ParseIP(input)

		fmt.Println("Enter Port to ping: ")
		fmt.Scanln(&input)
		port, _ := strconv.Atoi(input)

		fmt.Println("Enter message: ")
		fmt.Scanln(&input)
		message := input
		buffer := []byte(message)

		network.connection.WriteToUDP(buffer, &net.UDPAddr{IP:ip, Port:port, Zone:""})

		fmt.Println("Sent \"" + message + "\" to " + ip.String() + ":" + strconv.Itoa(port))
	}
}
