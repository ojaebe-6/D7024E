package main
import (
  "crypto/sha1"
  "hash"
)

type Kademlia struct {
  hash_table map[[20]byte][]byte
  sha hash.Hash
}

func NewKademlia() *Kademlia {
  return &Kademlia{hash_table:make(map[[20]byte][]byte), sha:sha1.New()}
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Kademlia) LookupData(hash [20]byte) []byte {
  if _, ok := kademlia.hash_table[hash]; ok {
    return kademlia.hash_table[hash]
  }
  return nil
}

func (kademlia *Kademlia) Store(data []byte) [20]byte {
  var hashed_data [20]byte
  copy(kademlia.sha.Sum(data)[:], hashed_data[0:20])
  kademlia.hash_table[hashed_data] = data
  return hashed_data
}
