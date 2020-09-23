package main

import (
	"fmt"
	"os"
	"net"
	"strconv"
)

func bootstrap(network *Network, nodeIPs []string) {
	successfulPing := false
	for _, stringIP := range nodeIPs {
		ip := net.ParseIP(stringIP)
		if ip != nil {
			fmt.Println("Pinging bootstrap node " + ip.String())
			contact := NewContact(nil, ip)
			success := network.SendPingMessage(&contact)

			if success {
				fmt.Println("Ping successful!");
				successfulPing = true
			} else {
				fmt.Println("Ping failed!");
			}
		}
	}

	if !successfulPing {
		fmt.Println("No bootstrap nodes answered. Bootstrap failed");
	} else {
		//TODO self lookup

		//Successful bootstrap
		fmt.Println("Bootstrap successful!");
	}
}

func main() {
	kademlia := NewKademlia(NewRandomKademliaID())
	network := NewNetwork(kademlia)

	fmt.Println("Node initalized on port " + strconv.Itoa(standardPort) + "!");

	bootstrap(network, os.Args)

	//Sleep forever, work done in goroutines
	select{}
}
