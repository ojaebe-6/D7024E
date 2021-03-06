package main
import (
  "crypto/sha1"
	"sync"
)

type Kademlia struct {
  hash_table map[[20]byte][]byte
  routing_table *RoutingTable
	routingTableMutex sync.RWMutex
	myID *KademliaID
}

func NewKademlia(id *KademliaID) *Kademlia {
  return &Kademlia{hash_table:make(map[[20]byte][]byte), routing_table:NewRoutingTable(id), myID:id}
}

func (kademlia *Kademlia) AddContact(contact *Contact) {
	kademlia.routingTableMutex.Lock()
	kademlia.routing_table.AddContact(*contact)
	kademlia.routingTableMutex.Unlock()
}

func (kademlia *Kademlia) AddContacts(contacts []Contact) {
	kademlia.routingTableMutex.Lock()
	for _, contact := range contacts {
			kademlia.routing_table.AddContact(contact)
	}
	kademlia.routingTableMutex.Unlock()
}

func (kademlia *Kademlia) LookupContact(target *KademliaID) []Contact {
	kademlia.routingTableMutex.RLock()
	contacts := kademlia.routing_table.FindClosestContacts(target, 20)
	kademlia.routingTableMutex.RUnlock()
	return contacts
}

func (kademlia *Kademlia) LookupData(hash [20]byte) []byte {
  if _, ok := kademlia.hash_table[hash]; ok {
    return kademlia.hash_table[hash]
  }
  return nil
}

func (kademlia *Kademlia) Store(data []byte) [20]byte {
  hashed_data := sha1.Sum(data)
  kademlia.hash_table[hashed_data] = append([]byte(nil), data...)
  return hashed_data
}
