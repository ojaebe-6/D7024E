package main

import (
	"log"
	"net"
)

type Network struct {
	connection *net.UDPConn
}

func Listen(ip string, port int) Network {
	connection, error := net.ListenUDP("udp", &net.UDPAddr{IP:net.ParseIP(ip), Port:port, Zone:""})
	if error != nil {
		log.Fatal(error)
	}

	return Network{connection}
}

/*
func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}*/
