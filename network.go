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
}

func NewNetwork() *Network {
	connection, error := net.ListenUDP("udp", &net.UDPAddr{IP:nil, Port:standardPort, Zone:""})
	if error != nil {
		log.Fatal(error)
	}

	network := Network{connection:connection, responses:[]*NetworkResponse{}}

	go func() {
		for {
			data := []byte{}
			_, senderAddress, error := connection.ReadFromUDP(data)

			if error != nil {
				log.Fatal(error)
			}

			if len(data) > 0 {
				network.handleNetworkData(senderAddress, data)
			}
		}
	}()

	return &network
}

func (network *Network) handleNetworkData(senderAddress *net.UDPAddr, data []byte) {
	if data[0] == networkVersion {
		messageType := data[1]
		magicValue := binary.LittleEndian.Uint64(data[2:10])

		if messageType == ResponsePing || messageType == ResponseStore || messageType == ResponseFindNode || messageType == ResponseFindValue {
			network.handleNetworkDataResponse(messageType, magicValue, data)
		} else {
			network.handleNetworkDataRequest(senderAddress, messageType, magicValue, data)
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
		response.contacts = dataToContacts(data[11:])
	case ResponseFindValue:
		isData := (data[11] == 1)
		if isData {
			response.data = data[12:]
		} else {
			response.contacts = dataToContacts(data[12:])
		}
	}
	response.mutex.Unlock()
}

func (network *Network) handleNetworkDataRequest(senderAddress *net.UDPAddr, messageType byte, magicValue uint64, data []byte) {
	switch messageType {
	case MessagePing:
		network.SendMessageResponse(senderAddress, ResponsePing, magicValue, func(buffer *bytes.Buffer){})
	case MessageStore:
		//TODO
	case MessageFindNode:
		//TODO
	case MessageFindValue:
		//TODO
	}
}

func dataToContacts(data []byte) []Contact {
	const ipAddressLength = 4
	contactDataLength := IDLength + ipAddressLength
	contacts := make([]Contact, len(data) / contactDataLength)

	for i := 0; i < len(data); i += contactDataLength {
		address := "0.0.0.0"//TODO
		id := NewKademliaIDFromBytes(data[i:i + IDLength])
		contacts[i] = NewContact(id, address)
	}

	return contacts
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

	//Write custom data
	writeData(buffer)

	//Send data
	network.connection.WriteToUDP(buffer.Bytes(), &net.UDPAddr{IP:net.ParseIP(contact.Address), Port:standardPort, Zone:""})

	//Add to response waiting list
	network.responsesMutex.Lock()
	response := NetworkResponse{magicValue:magicValue, answered:false}
	network.responses = append(network.responses, &response)
	network.responsesMutex.Unlock()

	//Await response or timeout
	startTime := time.Now()
	for time.Since(startTime).Milliseconds() >= responseTimeout {
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

func (network *Network) SendFindDataMessage(contact *Contact, hash string) (bool, []Contact, []byte) {
	response := network.SendMessage(contact, MessageFindValue, func(buffer *bytes.Buffer) {
		buffer.WriteString(hash)
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
