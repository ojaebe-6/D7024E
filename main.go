package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() {
	network := NewNetwork()

	fmt.Println("Node initalized on port 20000!");

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
