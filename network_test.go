package main

import (
	"crypto/sha1"
	"bytes"
	"fmt"
	"net"
	"testing"
)

func TestNetwork(t *testing.T) {
	selfID := NewKademliaID("1111111100000000000000000000000000000000")
	self := NewContact(selfID, net.ParseIP("127.0.0.1"))

	unreachable := NewContact(NewKademliaID("1111111200000000000000000000000000000000"), nil)

	kademlia := NewKademlia(selfID)
	network := NewNetwork(kademlia)

	//Ping
	if network.SendPingMessage(&self) != true {
		t.Error("Failed to self ping")
	}
	fmt.Println("Waiting for timeout...")
	if network.SendPingMessage(&unreachable) != false {
		t.Error("Pinged unreachable node")
	}

	//Find contact
	findContactSuccess, findContactContacts := network.SendFindContactMessage(&self, NewRandomKademliaID())
	if !findContactSuccess {
		t.Error("Failed to find contacts")
	}
	if len(findContactContacts) != 1 {
		t.Error("Incorrect number of contacts")
	} else if findContactContacts[0].String() != self.String() {
		t.Error("Incorrect contact")
	}

	//Store and find data
	data := []byte("Hello")
	hash := sha1.Sum(data)

	//Find data before stored
	findDataSuccess, findDataContacts, findDataData := network.SendFindDataMessage(&self, hash)
	if findDataSuccess != true {
		t.Error("Failed to request data")
	}
	if len(findDataContacts) != 1 {
		t.Error("Did not return contacts")
	}
	if len(findDataData) != 0 {
		t.Error("Found non-existant data")
	}

	//Store data
	if network.SendStoreMessage(&self, data) != true {
		t.Error("Failed to store")
	}

	//Find data after stored
	findDataSuccess2, findDataContacts2, findDataData2 := network.SendFindDataMessage(&self, hash)
	if findDataSuccess2 != true {
		t.Error("Failed to find data")
	}
	if len(findDataContacts2) != 0 {
		t.Error("Returned contacts")
	}
	if !bytes.Equal(findDataData2, data) {
		t.Error("Found incorrect data")
	}
}
