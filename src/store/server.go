package db

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Rosa-Devs/Database/src/manifest"
	gostream "github.com/libp2p/go-libp2p-gostream"
	p2phttp "github.com/libp2p/go-libp2p-http"
)

type MerkelRoot struct {
	Root string `json:"root"`
}

type RecordRequest struct {
	Id       string            `json:"id"`
	Database manifest.Manifest `json:"database"`
	Pool     string            `json:"pool"`
}

type RecordResponse struct {
	Data string `json:"data"`
}

func (r *RecordRequest) Decode(path string) {
	parts := strings.Split(path, "/")

	// Check if the link has the expected number of parts
	if len(parts) != 4 {
		log.Println("invalid link format")
	}

	// Check if the last part has the expected ".json" extension
	if !strings.HasSuffix(parts[3], ".json") {
		log.Println("invalid link format")
	}

	// Extract the components from the link
	r.Id = strings.TrimSuffix(parts[3], ".json")
	r.Pool = parts[1]
}
func (r *RecordRequest) Serialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

// Deserialize deserializes a JSON string to an Action struct.
func (r *RecordRequest) Deserialize(jsonData []byte) error {
	err := json.Unmarshal(jsonData, r)
	if err != nil {
		return err
	}
	return nil
}

func (r *RecordResponse) Serialize() ([]byte, error) {
	return json.Marshal(r)
}

// Deserialize deserializes a JSON string to an Action struct.
func (r *RecordResponse) Deserialize(jsonData []byte) error {
	err := json.Unmarshal(jsonData, r)
	if err != nil {
		return err
	}
	return nil
}

func (s *DB) Serve() {
	//Create clinet
	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", p2phttp.NewTransport(s.H))
	s.client = &http.Client{Transport: tr}

	listener, _ := gostream.Listen(s.H, p2phttp.DefaultP2PProtocol)
	defer listener.Close()
	//REGISTER SERVER HANDLERS
	http.HandleFunc("/merkle", s.merkleHandler)
	http.HandleFunc("/indexs", s.IndexHandler)
	http.HandleFunc("/get", s.GetRecord)

	//START SERVER
	s.server = &http.Server{}
	s.server.Serve(listener)
}

func (s *DB) merkleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	m := new(manifest.Manifest)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	m.Deserialize(body)

	db := s.GetDb(*m)
	root, err := db.GenereateMerkleTree()
	if err != nil {
		http.Error(w, "Fail to generete mekrle tree", http.StatusInternalServerError)
	}

	// Return the root hash in the response
	responseData := MerkelRoot{Root: root}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func (s *DB) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	m := new(manifest.Manifest)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	m.Deserialize(body)

	db := s.GetDb(*m)
	responseData, err := db.Index()
	if err != nil {
		http.Error(w, "Fail to generete mekrle tree", http.StatusInternalServerError)
	}

	// Return the root hash in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func (s *DB) GetRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	m := new(RecordRequest)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	m.Deserialize(body)

	db := s.GetDb(*&m.Database)

	pool, err := db.GetPool(m.Pool)
	if err != nil {
		http.Error(w, "Failed to get pool", http.StatusInternalServerError)
		return
	}

	record, err := pool.GetByID(m.Id)
	if err != nil {
		http.Error(w, "Failed get record", http.StatusInternalServerError)
	}

	//log.Println("Record", record)
	record_str, err := json.Marshal(record)
	if err != nil {
		http.Error(w, "Failed serialize record", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record_str)
}
