package main

import (
	"fmt"
	"net"
	"testing"
)

func TestRoutingTable(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), net.ParseIP("192.168.0.1")))

	rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), net.ParseIP("192.168.0.2")))
	rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), net.ParseIP("192.168.0.3")))
	rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), net.ParseIP("192.168.0.4")))
	rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), net.ParseIP("192.168.0.5")))
	rt.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000000"), net.ParseIP("192.168.0.6")))
	rt.AddContact(NewContact(NewKademliaID("2111111400000000000000000000000000000000"), net.ParseIP("192.168.0.7")))

	contacts := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 20)
	for i := range contacts {
		fmt.Println(contacts[i].String())
	}
}
