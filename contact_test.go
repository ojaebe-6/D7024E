package main

import (
	"net"
	"testing"
)

func TestContact(t *testing.T) {
	//Contact
	contact := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), net.ParseIP("192.168.0.1"))
	id := NewKademliaID("3333333300000000000000000000000000000000")
	contact.CalcDistance(id)
	if contact.distance.String() != "2222222200000000000000000000000000000000" {
		t.Error("Distance is incorrect")
	}

	contact2 := NewContact(NewKademliaID("4444444400000000000000000000000000000000"), net.ParseIP("192.168.0.2"))
	contact2.CalcDistance(id)
	if contact2.Less(&contact) {
		t.Error("Less() incorrect")
	}
	if !contact.Less(&contact2) {
		t.Error("Less() incorrect")
	}

	contactString := NewContact(NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc34"), net.ParseIP("192.168.0.1"))
	if contactString.String() != "contact(\"d406303f608bf7270f34dbd7c55d49cf767bbc34\", \"192.168.0.1\")" {
		t.Error("Contact.String() is incorrect (" + contactString.String() + ")")
	}

	//ContactCandidates
	var candidates ContactCandidates

	candidates.Append([]Contact{contact})

	if len(candidates.GetContacts(0)) != 0 {
		t.Error("GetContacts(0) returned contacts")
	}
	if len(candidates.GetContacts(1)) != 1 {
		t.Error("GetContacts(1) did not return 1 contact")
	}
	if candidates.GetContacts(1)[0].String() != "contact(\"1111111100000000000000000000000000000000\", \"192.168.0.1\")" {
		t.Error("GetContacts(1)[0].String() incorrect string (" + candidates.GetContacts(1)[0].String() + ")")
	}

	if candidates.Len() != 1 {
		t.Error("Len() is not 1")
	}

	//TODO Sort, Swap, Less
}
