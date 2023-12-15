package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Rosa-Devs/Database/src/manifest"

	db "github.com/Rosa-Devs/Database/src/store"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "pubsub-chat-example"

func main() {
	database := flag.String("d", "", "use it to create databse manifest file")
	ManifestFile := flag.String("m", "", "set Manifets file")
	FolderName := flag.String("f", "", "set db folder name")
	NickName := flag.String("n", "", "Set nickname")
	flag.Parse()

	if *database != "" {
		manifest.GenereateManifest(*database, true)
		return
	}

	if *ManifestFile == "" {
		log.Println("Specifi a manifest file... -m")
		return
	}

	if *NickName == "" {
		log.Println("Specifi a nickname!")
		return
	}

	ctx := context.Background()

	// create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New()
	if err != nil {
		panic(err)
	}

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	// setup local mDNS discovery
	if err := setupDiscovery(h); err != nil {
		panic(err)
	}

	// use the nickname from the cli flag, or a default if blank

	manifetstData := manifest.ReadManifestFromFile(*ManifestFile)

	// !! GLOBAl DB MANAGER !!
	//CREATE DATABSE INSTANCE
	Drvier := db.DB{
		H:  h,
		Pb: ps,
	}
	//START DATABSE INSTANCE
	if *FolderName != "" {
		Drvier.Start(*FolderName)
	} else {
		Drvier.Start("test_db_1")
	}
	//CREATE TEST DB
	Drvier.CreateDb(manifetstData)

	// !! WORKING WITH SPECIFIED BATABASE !!
	db1 := Drvier.GetDb(manifetstData)
	db1.StartWorker(15)

	err = db1.CreatePool("chat")
	if err != nil {
		log.Println("Mayby this pool alredy exist:", err)
		//return
	}

	pool, err := db1.GetPool("chat")
	if err != nil {
		log.Println(err)
		return
	}

	updateListener := make(chan db.Event)
	db1.EventBus.Subscribe(db.DbUpdateEvent, updateListener)

	go func() {
		var prevState []map[string]interface{}
		for {
			for _ = range updateListener {
				filter := map[string]interface{}{
					"type": 1,
				}

				data, err := pool.Filter(filter)
				if err != nil {
					fmt.Println("Data:", data)
					fmt.Println("Error filtering data:", err)
				}

				// Sort messages by timestamp in descending order (newest first)
				sort.Slice(data, func(i, j int) bool {
					time1, _ := time.Parse(time.RFC3339, data[i]["TimeStamp"].(string))
					time2, _ := time.Parse(time.RFC3339, data[j]["TimeStamp"].(string))
					return time1.After(time2)
				})

				for _, record := range data {
					if isNewMessage(record, prevState, *NickName) {
						fmt.Println(record["nick"].(string) + ":" + record["msg"].(string))
					}
				}

				prevState = data
				time.Sleep(time.Millisecond * 70)
			}
		}
	}()

	reader := bufio.NewReader(os.Stdout)

	for {
		// Read input from the user
		fmt.Print(*NickName + ":")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		// Trim spaces and newline characters
		text = text[:len(text)-1]

		if len(text) == 0 {
			continue
		}
		// Print the input
		msg := new(Message)
		msg.Nick = *NickName
		msg.Msg = text
		msg.Type = TYPE_MSG
		msg.TimeStamp = time.Now()

		json_data, err := msg.Serialize()
		if err != nil {
			log.Println(err)
		}
		err = pool.Record(json_data)
		//fmt.Println("You entered:", text)
	}

}

func isNewMessage(msg map[string]interface{}, lastState []map[string]interface{}, nick string) bool {
	for _, lastMsg := range lastState {
		lastTimeStamp, lastTimeStampOK := lastMsg["TimeStamp"].(string)
		currentTimeStamp, currentTimeStampOK := msg["TimeStamp"].(string)

		if !lastTimeStampOK || !currentTimeStampOK {
			// Handle error case where timestamp is not a valid string
			continue
		}

		time1, err1 := time.Parse(time.RFC3339, lastTimeStamp)
		time2, err2 := time.Parse(time.RFC3339, currentTimeStamp)

		if err1 != nil || err2 != nil {
			// Handle error case where parsing timestamp fails
			continue
		}

		if time1.Equal(time2) && msg["Nick"] == lastMsg["Nick"] && msg["Msg"] == lastMsg["Msg"] {
			return false
		}

		if msg["nick"] == nick {
			return false
		}
	}

	return true
}

const TYPE_MSG = 1

type Message struct {
	Type      int `json:"type"`
	TimeStamp time.Time
	Nick      string `json:"nick"`
	Msg       string `json:"msg"`
}

func (a *Message) Serialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func (a *Message) Deserialize(jsonDaat []byte) error {
	err := json.Unmarshal(jsonDaat, a)
	if err != nil {
		return err
	}
	return nil
}

func getInputString(prompt string) string {
	fmt.Print(prompt + " ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		os.Exit(1)
	}
	return strings.TrimSpace(input)
}

// MDNS
type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID)
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID, err)
	}
}

func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}
