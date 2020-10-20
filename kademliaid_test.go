package main

import (
	"testing"
)

func TestKademliaID(t *testing.T) {
	hexLowerCase := NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc34")
	if hexLowerCase.String() != "d406303f608bf7270f34dbd7c55d49cf767bbc34" {
		t.Error("Hex lower case incorrect")
	}

	hexUpperCase := NewKademliaID("D406303F608BF7270F34DBD7C55D49CF767BBC34")
	if hexUpperCase.String() != "d406303f608bf7270f34dbd7c55d49cf767bbc34" {
		t.Error("Hex upper case incorrect")
	}

	var bytes [20]byte
	for i := 0; i < 20; i++ {
		bytes[i] = 255
	}
	fromBytesID := NewKademliaIDFromBytes(bytes[:])
	if fromBytesID.String() != "ffffffffffffffffffffffffffffffffffffffff" {
		t.Error("NewKademliaIDFromBytes() failed to produce correct ID, incorrect ID: " + fromBytesID.String())
	}

	randomA := NewRandomKademliaID()
	randomB := NewRandomKademliaID()
	if randomA.String() == randomB.String() {
		t.Error("Random IDs are equal")
	}

	low := NewKademliaID("1111111100000000000000000000000000000000")
	high := NewKademliaID("3333333300000000000000000000000000000000")
	if !low.Less(high) {
		t.Error("Lower ID is not less than high ID")
	}
	if high.Less(low) {
		t.Error("High ID is less than low ID")
	}

	if low.CalcDistance(high).String() != "2222222200000000000000000000000000000000" {
		t.Error("Distance is incorrect")
	}
	if low.CalcDistance(high).String() != high.CalcDistance(low).String() {
		t.Error("Distance is not symmetrical")
	}

	equalA := NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc34")
	equalB := NewKademliaID("d406303f608bf7270f34dbd7c55d49cf767bbc34")
	if !equalA.Equals(equalB) {
		t.Error("Equal IDs are not equal")
	}
}
