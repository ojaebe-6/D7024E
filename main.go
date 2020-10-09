package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

func sortContactsByTargetDistance(contacts []Contact, target *KademliaID) {
	sort.SliceStable(contacts, func(i, j int) bool {
		return contacts[i].ID.CalcDistance(target).Less(contacts[j].ID.CalcDistance(target))
	})
}

func LookupContact(kademlia *Kademlia, network *Network, target *KademliaID, maxCount int) []Contact {
	simultaneousLookups := 3
	maxLookupsSinceBestFound := 6

	uniqueContacts := make(map[KademliaID]bool)

	//Local lookup
	contacts := kademlia.LookupContact(target);
	if len(contacts) == 0 {
		return nil
	}
	sortContactsByTargetDistance(contacts, target)

	allContacts := make([]Contact, len(contacts))
	copy(allContacts, contacts)

	for _, contact := range contacts {
		uniqueContacts[*contact.ID] = true
	}

	mutex := sync.Mutex{}
	bestContact := contacts[0]
	lookupsSinceBestFound := 0
	currentLookups := 0

	for i := 0; i < simultaneousLookups; i++ {
		go func() {
			for {
				mutex.Lock()
				if lookupsSinceBestFound > maxLookupsSinceBestFound || (len(contacts) == 0 && currentLookups == 0) {
					mutex.Unlock()
					break
				}

				if len(contacts) > 0 {
					contact := contacts[0]

					//Remove first element
					contacts = contacts[1:]

					currentLookups++
					mutex.Unlock()

					success, newContacts := network.SendFindContactMessage(&contact, target)
					mutex.Lock()
					currentLookups--
					if success {
						for _, contact := range newContacts {
							_, exists := uniqueContacts[*contact.ID]
							if !exists {
								contacts = append(contacts, contact)
								allContacts = append(allContacts, contact)
								uniqueContacts[*contact.ID] = true
							}
						}
						sortContactsByTargetDistance(contacts, target)

						if len(contacts) > 0 {
							if contacts[0].ID.CalcDistance(target).Less(bestContact.ID.CalcDistance(target)) {
								bestContact = contacts[0]
								if lookupsSinceBestFound > maxLookupsSinceBestFound {
									mutex.Unlock()
									break
								} else {
									lookupsSinceBestFound = 0
								}
							}
						}
						lookupsSinceBestFound++
					}
					mutex.Unlock()
				} else {
					mutex.Unlock()
				}
				time.Sleep(1)
			}
		}()
	}

	//Wait until go routines are finished
	for {
		mutex.Lock()
		if lookupsSinceBestFound > maxLookupsSinceBestFound || (len(contacts) == 0 && currentLookups == 0) {
			break
		}
		mutex.Unlock()
		time.Sleep(1)
	}

	sortContactsByTargetDistance(allContacts, target)

	if len(allContacts) < maxCount {
		maxCount = len(allContacts)
	}
	return allContacts[:maxCount]
}

func LookupData(kademlia *Kademlia, network *Network, hash [20]byte) []byte {
	//Local data lookup
	data := kademlia.LookupData(hash)
	if data != nil {
		return data
	}

	target := NewKademliaIDFromBytes(hash[:])

	simultaneousLookups := 3
	maxLookupsSinceBestFound := 6

	uniqueContacts := make(map[KademliaID]bool)

	//Local node lookup
	contacts := kademlia.LookupContact(target);
	if len(contacts) == 0 {
		return nil
	}
	sortContactsByTargetDistance(contacts, target)

	for _, contact := range contacts {
		uniqueContacts[*contact.ID] = true
	}

	mutex := sync.Mutex{}
	bestContact := contacts[0]
	lookupsSinceBestFound := 0
	currentLookups := 0

	for i := 0; i < simultaneousLookups; i++ {
		go func() {
			for {
				mutex.Lock()
				if data != nil || lookupsSinceBestFound > maxLookupsSinceBestFound || (len(contacts) == 0 && currentLookups == 0) {
					mutex.Unlock()
					break
				}

				if len(contacts) > 0 {
					contact := contacts[0]

					//Remove first element
					contacts = contacts[1:]

					currentLookups++
					mutex.Unlock()

					success, newContacts, newData := network.SendFindDataMessage(&contact, hash)
					mutex.Lock()
					currentLookups--
					if success {
						if len(newData) > 0 {
							if data == nil {
								data = newData
							}
							mutex.Unlock()
							break
						}

						for _, contact := range newContacts {
							_, exists := uniqueContacts[*contact.ID]
							if !exists {
								contacts = append(contacts, contact)
								uniqueContacts[*contact.ID] = true
							}
						}
						sortContactsByTargetDistance(contacts, target)

						if len(contacts) > 0 {
							if contacts[0].ID.CalcDistance(target).Less(bestContact.ID.CalcDistance(target)) {
								bestContact = contacts[0]
								if lookupsSinceBestFound > maxLookupsSinceBestFound {
									mutex.Unlock()
									break
								} else {
									lookupsSinceBestFound = 0
								}
							}
						}
						lookupsSinceBestFound++
					}
					mutex.Unlock()
				} else {
					mutex.Unlock()
				}
				time.Sleep(1)
			}
		}()
	}

	//Wait until go routines are finished
	for {
		mutex.Lock()
		if data != nil || lookupsSinceBestFound > maxLookupsSinceBestFound || (len(contacts) == 0 && currentLookups == 0) {
			break
		}
		mutex.Unlock()
		time.Sleep(1)
	}

	return data
}

func StoreData(kademlia *Kademlia, network *Network, data []byte, replicationFactor int) [20]byte {
	var hash [20]byte
	copy(hash[0:20], kademlia.sha.Sum(data)[:])

	target := NewKademliaIDFromBytes(hash[:])
	contacts := LookupContact(kademlia, network, target, replicationFactor)

	if len(contacts) == 0 {
		kademlia.Store(data)
	} else {
		if len(contacts) < replicationFactor {
			replicationFactor = len(contacts)
		}

		for i := 0; i < replicationFactor; i++ {
			go func(contact *Contact) {
				network.SendStoreMessage(contact, data)
			}(&contacts[i])
		}
	}

	return hash
}

func bootstrap(kademlia *Kademlia, network *Network, nodeIPs []string) {
	successfulPing := false
	for _, stringIP := range nodeIPs {
		ip := net.ParseIP(stringIP)
		if ip != nil {
			fmt.Println("Pinging bootstrap node " + ip.String())
			contact := NewContact(nil, ip)
			success := network.SendPingMessage(&contact)

			if success {
				fmt.Println("Ping successful!");
				successfulPing = true
			} else {
				fmt.Println("Ping failed!");
			}
		}
	}

	if !successfulPing {
		fmt.Println("No bootstrap nodes answered. Bootstrap failed");
	} else {
		//Lookup my ID and add contacts to k-buckets
		kademlia.AddContacts(LookupContact(kademlia, network, kademlia.myID, bucketSize))

		//Successful bootstrap
		fmt.Println("Bootstrap successful!");
	}
}

func main() {
	kademlia := NewKademlia(NewRandomKademliaID())
	network := NewNetwork(kademlia)

	fmt.Println("Node " + kademlia.myID.String() + " initalized on port " + strconv.Itoa(standardPort) + "!");

	bootstrap(kademlia, network, os.Args)
	CmdInterface(kademlia, network);
}

func Put(content string, kademlia *Kademlia, network *Network) [20]byte{
	var hash [20]byte;
	dataToSend := []byte(content);
	// Returned Hash value stores in byte slice.
	hash = StoreData(kademlia, network, dataToSend,5 );
	fmt.Println(hex.EncodeToString(hash[:]));
	return hash;
}

func Get(convertHash string , kademlia *Kademlia, network *Network, isLocal bool){
	var outStr []byte;
	var outFixed [20]byte;
	if(len(convertHash) == 40){
		outData, err := hex.DecodeString(convertHash);
		if err != nil {
			fmt.Println("Failed to get data.");
		}
		copy(outFixed[0:20], outData[:]);
		if(isLocal){
			outStr = kademlia.LookupData(outFixed);

		}else{
			outStr = LookupData(kademlia, network, outFixed);
		}
		if(outStr == nil){
			fmt.Println("Empty Data.");
		} else{
			fmt.Println(string(outStr));
		}
		// fmt.Println("%convert\n", outData);
	} else{
		fmt.Println("hex string is not 40 in length.");
	}
}

func CmdInterface(kademlia *Kademlia, network *Network) {
	var in string;
	lines := "----------------------------------------";
	var textIn string;

	for (in != "exit") {
		fmt.Println(lines);
		fmt.Println("| Node: TEST.\t       |");
		fmt.Println(lines);
		fmt.Println("| Options: put | get | getlocal | exit |");
		fmt.Println(lines);
		fmt.Print("| Option: ");
		fmt.Scanln(&in);
		switch in {
		case "put":
			// Store to byte array
			fmt.Println();
			fmt.Print("| Add data to be stored: ");
			fmt.Scanln(&textIn);
			// because golang just simply works like this, we assign string input to a
			// byte slice, as for replication Factor, no idea how many copies we want to store.
			// As for where to use this function, beats me.
			fmt.Print("| Hash: ")
			Put(textIn, kademlia, network);
			fmt.Println();
			continue;
		case "get":
			fmt.Print("| Hash for data: ");
			fmt.Scanln(&textIn);
			{
				fmt.Print("| Output: ")
				Get(textIn, kademlia, network, false);
				fmt.Println();
			}
			continue;

		case "getlocal":
			fmt.Print("| Hash for local data: ");
			fmt.Scanln(&textIn);
			{
				fmt.Print("| Output: ")
				Get(textIn, kademlia, network, true);
				fmt.Println();
			}
			continue;

		case "exit":
			fmt.Println("Exit");
			continue;
		}
	}
}