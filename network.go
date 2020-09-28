package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"log"
	"math"
	"math/big"
	"net"
	"sync"
	"time"
)

const networkVersion = 0;
const standardPort = 20000
const responseTimeout = 10000

type NetworkResponse struct {
	answered bool
	mutex sync.RWMutex

	magicValue uint64
	contacts []Contact
	data []byte
}

type Network struct {
	connection *net.UDPConn
	responses []*NetworkResponse
	responsesMutex sync.RWMutex
	kademlia *Kademlia
}

func NewNetwork(kademlia *Kademlia) *Network {
	connection, error := net.ListenUDP("udp", &net.UDPAddr{IP:nil, Port:standardPort, Zone:""})
	if error != nil {
		log.Fatal(error)
	}

	network := Network{connection:connection, responses:[]*NetworkResponse{}, kademlia:kademlia}

	data := make([]byte, 4096)
	go func() {
		for {
			length, senderAddress, error := connection.ReadFromUDP(data)

			if error != nil {
				log.Fatal(error)
			}

			if len(data) > 0 {
				network.handleNetworkData(senderAddress, data[:length])
			}
		}
	}()

	return &network
}

func (network *Network) handleNetworkData(senderAddress *net.UDPAddr, data []byte) {
	if data[0] == networkVersion {
		messageType := data[1]
		magicValue := binary.LittleEndian.Uint64(data[2:10])

		//Add contact to routing table
		id := NewKademliaIDFromBytes(data[10:10 + IDLength])
		contact := NewContact(id, senderAddress.IP)
		network.kademlia.AddContact(&contact)

		if messageType == ResponsePing || messageType == ResponseStore || messageType == ResponseFindNode || messageType == ResponseFindValue {
			network.handleNetworkDataResponse(messageType, magicValue, data[10 + IDLength:])
		} else {
			network.handleNetworkDataRequest(senderAddress, messageType, magicValue, data[10 + IDLength:])
		}
	}
}

func (network *Network) handleNetworkDataResponse(messageType byte, magicValue uint64, data []byte) {
	var response *NetworkResponse = nil
	network.responsesMutex.RLock()
	for _, response2 := range network.responses {
		if response2.magicValue == magicValue {
			response = response2
			break
		}
	}
	network.responsesMutex.RUnlock()

	response.mutex.Lock()
	response.answered = true
	switch messageType {
	case ResponseFindNode:
		response.contacts = dataToContacts(data)
	case ResponseFindValue:
		isData := (data[0] == 1)
		if isData {
			response.data = data[1:]
		} else {
			response.contacts = dataToContacts(data[1:])
		}
	}
	response.mutex.Unlock()
}

func (network *Network) handleNetworkDataRequest(senderAddress *net.UDPAddr, messageType byte, magicValue uint64, data []byte) {
	switch messageType {
	case MessagePing:
		network.SendMessageResponse(senderAddress, ResponsePing, magicValue, func(buffer *bytes.Buffer){})
	case MessageStore:
		network.kademlia.Store(data)
		network.SendMessageResponse(senderAddress, ResponseStore, magicValue, func(buffer *bytes.Buffer){})
	case MessageFindNode:
		contacts := network.kademlia.LookupContact(NewKademliaIDFromBytes(data))
		network.SendMessageResponse(senderAddress, ResponseFindNode, magicValue, func(buffer *bytes.Buffer) {
			buffer.Write(contactsToData(contacts))
		})
	case MessageFindValue:
		var hash [20]byte
		copy(hash[:], data[0:20])
		data := network.kademlia.LookupData(hash)
		if data == nil {
			contacts := network.kademlia.LookupContact(NewKademliaIDFromBytes(hash[:]))
			network.SendMessageResponse(senderAddress, ResponseFindValue, magicValue, func(buffer *bytes.Buffer) {
				buffer.WriteByte(0)
				buffer.Write(contactsToData(contacts))
			})
		} else {
			network.SendMessageResponse(senderAddress, ResponseFindValue, magicValue, func(buffer *bytes.Buffer) {
				buffer.WriteByte(1)
				buffer.Write(data)
			})
		}
	}
}

func dataToContacts(data []byte) []Contact {
	const ipAddressLength = 4
	contactDataLength := IDLength + ipAddressLength
	contacts := make([]Contact, len(data) / contactDataLength)

	for i := 0; i < len(data); i += contactDataLength {
		id := NewKademliaIDFromBytes(data[i:i + IDLength])
		address := net.IPv4(data[i + IDLength + 0], data[i + IDLength + 1], data[i + IDLength + 2], data[i + IDLength + 3])
		contacts[i] = NewContact(id, address)
	}

	return contacts
}

func contactsToData(contacts []Contact) []byte {
	const ipAddressLength = 4
	contactDataLength := IDLength + ipAddressLength
	data := make([]byte, len(contacts) * contactDataLength)

	for i := 0; i < len(contacts); i++ {
		//Copy ID
		for j := 0; j < IDLength; j++ {
			data[i * contactDataLength + j] = contacts[i].ID[j]
		}

		//Copy IP
		data[i * contactDataLength + IDLength + 0] = contacts[i].Address[12]
		data[i * contactDataLength + IDLength + 1] = contacts[i].Address[13]
		data[i * contactDataLength + IDLength + 2] = contacts[i].Address[14]
		data[i * contactDataLength + IDLength + 3] = contacts[i].Address[15]
	}

	return data
}

func (network *Network) SendMessageResponse(address *net.UDPAddr, messageType byte, magicValue uint64, writeData func(*bytes.Buffer)) {
	buffer := new(bytes.Buffer)

	//Version
	buffer.WriteByte(networkVersion)

	//Message type
	buffer.WriteByte(messageType)

	//Magic value
	error := binary.Write(buffer, binary.LittleEndian, magicValue)
	if error != nil {
		log.Fatal(error)
	}

	//My ID
	buffer.Write(network.kademlia.myID[:])

	//Write custom data
	writeData(buffer)

	//Send data
	network.connection.WriteToUDP(buffer.Bytes(), address)
}

func (network *Network) SendMessage(contact *Contact, messageType byte, writeData func(*bytes.Buffer)) *NetworkResponse {
	buffer := new(bytes.Buffer)

	//Version
	buffer.WriteByte(networkVersion)

	//Message type
	buffer.WriteByte(messageType)

	//Magic value
	magicValueBig, error := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if error != nil {
		log.Fatal(error)
	}
	magicValue := magicValueBig.Uint64()

	error2 := binary.Write(buffer, binary.LittleEndian, magicValue)
	if error2 != nil {
		log.Fatal(error2)
	}

	//My ID
	buffer.Write(network.kademlia.myID[:])

	//Write custom data
	writeData(buffer)

	//Send data
	network.connection.WriteToUDP(buffer.Bytes(), &net.UDPAddr{IP:contact.Address, Port:standardPort, Zone:""})

	//Add to response waiting list
	network.responsesMutex.Lock()
	response := NetworkResponse{magicValue:magicValue, answered:false}
	network.responses = append(network.responses, &response)
	network.responsesMutex.Unlock()

	//Await response or timeout
	startTime := time.Now()
	for time.Since(startTime).Milliseconds() < responseTimeout {
		response.mutex.RLock()
		if response.answered {
			response.mutex.RUnlock()
			break
		}
		response.mutex.RUnlock()

		time.Sleep(1)
	}

	//Remove from response waiting list
	network.responsesMutex.Lock()
	for index, response := range network.responses {
		if response.magicValue == magicValue {
			network.responses[index] = network.responses[len(network.responses) - 1]
			network.responses = network.responses[:len(network.responses) - 1]
			break
		}
	}
	network.responsesMutex.Unlock()

	if response.answered {
		return &response
	} else {
		//Timeout
		return nil
	}
}

func (network *Network) SendPingMessage(contact *Contact) bool {
	response := network.SendMessage(contact, MessagePing, func(buffer *bytes.Buffer){})

	if response != nil {
		//Success
		return true
	} else {
		//Fail
		return false
	}
}

func (network *Network) SendFindContactMessage(contact *Contact, id *KademliaID) (bool, []Contact) {
	response := network.SendMessage(contact, MessageFindNode, func(buffer *bytes.Buffer) {
		for _, b := range id {
			buffer.WriteByte(b)
		}
	})

	if response != nil {
		//Success
		return true, response.contacts
	} else {
		//Fail
		return false, []Contact{}
	}
}

func (network *Network) SendFindDataMessage(contact *Contact, hash [20]byte) (bool, []Contact, []byte) {
	response := network.SendMessage(contact, MessageFindValue, func(buffer *bytes.Buffer) {
		buffer.Write(hash[:])
	})

	if response != nil {
		if len(response.contacts) > 0 {
			//Return closer contacts
			return true, response.contacts, []byte{}
		} else {
			//Return stored data
			return true, []Contact{}, response.data
		}
	} else {
		//Fail
		return false, []Contact{}, []byte{}
	}
}

func (network *Network) SendStoreMessage(contact *Contact, data []byte) bool {
	response := network.SendMessage(contact, MessageStore, func(buffer *bytes.Buffer) {
		buffer.Write(data)
	})

	if response != nil {
		//Success
		return true
	} else {
		//Fail
		return false
	}
}
