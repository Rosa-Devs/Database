package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Rosa-Devs/Database/src/manifest"

	db "github.com/Rosa-Devs/Database/src/store"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type Manager struct {
	Dbs map[string]*db.Database
}

func (m *Manager) Add(db *db.Database, name string) {
	m.Dbs[name] = db
}

func (m *Manager) Get(name string) *db.Database {
	return m.Dbs[name]
}

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "pubsub-chat-example"

func main() {
	database := flag.String("d", "", "use it to create databse manifest file")
	ManifestFile := flag.String("m", "", "set Manifets file")
	FolderName := flag.String("f", "", "set db folder name")
	flag.Parse()

	if *database != "" {
		manifest.GenereateManifest(*database, true)
		return
	}

	if *ManifestFile == "" {
		log.Println("Specifi a manifest file... -m")
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

	err = db1.CreatePool("test_pool")
	if err != nil {
		log.Println("Mayby this pool alredy exist:", err)
		//return
	}

	_, err = db1.GetPool("test_pool")
	if err != nil {
		log.Println(err)
		return
	}

	for {
	}

}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID)
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID, err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}
