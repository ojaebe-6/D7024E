package main

import (
	"net"
	"testing"
)

func TestRoutingTable(t *testing.T) {
	rt := NewRoutingTable(NewKademliaID("FFFFFFFF00000000000000000000000000000000"))

	rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), net.ParseIP("192.168.0.2")))
	rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), net.ParseIP("192.168.0.3")))
	rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), net.ParseIP("192.168.0.4")))
	rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), net.ParseIP("192.168.0.5")))
	rt.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000000"), net.ParseIP("192.168.0.6")))
	rt.AddContact(NewContact(NewKademliaID("2111111400000000000000000000000000000000"), net.ParseIP("192.168.0.7")))

	contactsEmpty := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 0)
	if len(contactsEmpty) != 0 {
		t.Error("Returned contacts when requesting zero contacts")
	}

	contacts1 := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 1)
	if contacts1[0].ID.String() != "2111111400000000000000000000000000000000" {
		t.Error("Returned incorrect contact")
	}

	contacts2 := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 2)
	if contacts2[0].ID.String() != "2111111400000000000000000000000000000000" && contacts2[1].ID.String() != "1111111400000000000000000000000000000000" {
		t.Error("Returned incorrect contacts")
	}

	contacts20 := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 20)
	if len(contacts20) != 6 {
		t.Error("Incorrect length")
	}

	if rt.getBucketIndex(NewKademliaID("0000000000000000000000000000000000000000")) != 0 {
		t.Error("Incorrect bucket")
	}
	if rt.getBucketIndex(NewKademliaID("AAAAAAAA00000000000000000000000000000000")) != 1 {
		t.Error("Incorrect bucket")
	}
	if rt.getBucketIndex(NewKademliaID("FFFFFFFF00000000000000000000000000000000")) != 159 {
		t.Error("Incorrect bucket")
	}
}
