package main

import (
	"net"
	"bytes"
	"encoding/hex"
	"testing"
)

func TestKademlia(t *testing.T) {

	AddContactTest := NewKademlia(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc34"))

	contact := NewContact(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc35"), net.ParseIP("192.168.0.2"))

	AddContactTest.AddContact(&contact)
	if AddContactTest.routing_table.FindClosestContacts(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc35"),1)[0].String() != contact.String() {
		t.Error("Contact was not added")
	}

	AddContactsTest := NewKademlia(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc34"))

	contact1 := NewContact(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc35"), net.ParseIP("192.168.0.2"))
	contact2 := NewContact(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc36"), net.ParseIP("192.168.0.3"))
	contact3 := NewContact(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc37"), net.ParseIP("192.168.0.4"))

	AddContactsTest.AddContacts([]Contact{contact1, contact2, contact3})
	if((AddContactsTest.routing_table.FindClosestContacts(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc35"), 1)[0].String() != contact1.String()) &&
		 (AddContactsTest.routing_table.FindClosestContacts(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc36"), 1)[0].String() != contact2.String()) &&
		 (AddContactsTest.routing_table.FindClosestContacts(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc37"), 1)[0].String() != contact3.String())) {
		t.Error("Contacts were not added")
	}

	LookupContactTest := NewKademlia(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc34"))

	LookupContactTest.AddContact(&contact)
	if(LookupContactTest.LookupContact(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc35"))[0].String() != contact.String()) {
		t.Error("Contact could not be found")
	}

	StoreTest := NewKademlia(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc34"))

	StoreTest.Store([]byte("hej"))
	var decodedData [20]byte;
	data,_ := hex.DecodeString("c412b37f8c0484e6db8bce177ae88c5443b26e92")
	copy(decodedData[0:20], data[:]);
	if !bytes.Equal(StoreTest.hash_table[decodedData], []byte("hej")) {
		t.Error("Could not find data")
	}

	LookupDataTest := NewKademlia(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc34"))

	LookupDataTest.Store([]byte("hej"))
	if !bytes.Equal(LookupDataTest.LookupData(decodedData), []byte("hej")) {
		t.Error("Could not find data")
	}
}
