package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

const maxRetries = 3
const retryDelay = time.Second * 5

func (wr *WorkerRoom) Sync() {
	for {
		time.Sleep(time.Second * time.Duration(wr.timeout))
		//GET LIST OF NODES IN SUBSCRIPTION

		nodes := wr.topic.ListPeers()

		//log.Println("ALL NODES:", nodes)

		if len(nodes) == 0 {
			log.Println("No peers available NOT SYNCING")
			return
		}

		if len(nodes) > 15 {
			nodes = getRandomNodes(nodes, 15)
		}

		m_data := wr.db.manifest
		m_data.Chiper = "9756707289479916212080576755FYou"

		m, err := m_data.Serialize()
		// Send a POST request to each node
		if err != nil {
			log.Println("Err:", err)
		}
		var roots []string
		for _, id := range nodes {

			resp, err := wr.db.db.client.Post("libp2p://"+id.String()+"/merkle", "application/json", bytes.NewBuffer(m))
			if err != nil {
				log.Println("Failed to post manifest data to node:", id.String(), "Err:", err)
				continue
			}
			defer resp.Body.Close()

			// Read the response body
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error reading response body:", err)
			} else {

				var merkelResponse MerkelRoot
				err = json.Unmarshal(body, &merkelResponse)
				if err != nil {
					log.Println("Error decoding JSON response:", err)
					continue
				}
				//log.Printf("Response from node %s: %s\n", id.String(), body)
				roots = append(roots, merkelResponse.Root)
			}
		}

		root_hash := findMostRepeatedString(roots)
		if root_hash == "" {
			log.Println("DataBase is empty NOT SYNCING")
			return
		}
		//log.Println("TRUE_HASH:", root_hash)

		my_root, err := wr.db.GenereateMerkleTree()
		if err != nil {
			log.Println("Error generating merkle tree:", err)
		}

		if my_root == root_hash {
			log.Println("DB sync complete")
			return
		}

		var DBIndexs []map[string]string

		for _, id := range nodes {
			resp, err := wr.db.db.client.Post("libp2p://"+id.String()+"/indexs", "application/json", bytes.NewBuffer(m))
			if err != nil {
				log.Println("Failed to post manifest data to node:", id.String(), "Err:", err)
				continue
			}
			defer resp.Body.Close()

			var Index map[string]string
			err = json.NewDecoder(resp.Body).Decode(&Index)
			if err != nil {
				log.Println("Failed to decode indexs, ERR:", err)
				continue
			}
			//log.Printf("Response from node %s: %s\n", id.String(), body)
			DBIndexs = append(DBIndexs, Index)

		}
		index := mostRepeatedMap(DBIndexs)

		my_index, err := wr.db.Index()
		if err != nil {
			log.Println("Failt to index database:", err)
		}

		changed_file := wr.db.CalculateChangedFiles(index, my_index)

		for _, file := range changed_file {
			wr.GetRecordUpdate(file, nodes)
		}

		log.Println("DB sync complete")
		time.Sleep(time.Second * 60)
	}
}

func (wr *WorkerRoom) GetRecordUpdate(file string, nodes []peer.ID) {
	payload := new(RecordRequest)

	payload.Decode(file)
	payload.Database = wr.db.manifest

	data_file, err := payload.Serialize()
	if err != nil {
		log.Println("Failed to serialize, ERROR:", err.Error())

	}

	var data [][]byte
	for _, id := range nodes {
		resp, err := wr.fetchDataWithRetry(id.String(), data_file)
		if err != nil {
			log.Println("Failed to post manifest data to node:", id.String())

		}

		// Read the response body
		if err != nil {
			log.Println("Error reading response body:", err)
		} else {
			var deserializedBytes []byte
			err = json.Unmarshal(resp, &deserializedBytes)
			if err != nil {
				fmt.Println("Error unmarshaling to bytes:", err)

			}
			data = append(data, deserializedBytes)
		}
		time.Sleep(time.Millisecond * 100)
	}

	// Choose the best matching result based on your logic
	bestMatch := findMostRepeatedBytes(data)
	if bestMatch == nil {
		log.Println("No matching data found for file:", file)
	}
	pool, err := wr.db.GetPool(payload.Pool, true)
	if err != nil {
		log.Println("Failed to get pool:", err)
	}

	err = pool.RecordWithID(bestMatch, payload.Id)
	if err != nil {
		log.Println("Fail to update file", err)
	}
	log.Println("Updated Recod:", payload.Id)
}

func (wr *WorkerRoom) fetchDataWithRetry(id string, data_file []byte) ([]byte, error) {

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := wr.db.db.client.Post("libp2p://"+id+"/get", "application/json", bytes.NewBuffer(data_file))
		if err != nil {
			log.Printf("Failed to post manifest data to node %s (Attempt %d/%d): %v", id, attempt, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading response body:", err)
			return nil, err
		}

		return body, nil
	}

	return nil, fmt.Errorf("Failed after %d retries", maxRetries)
}

func getRandomNodes(nodes []peer.ID, count int) []peer.ID {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	if count > len(nodes) {
		count = len(nodes)
	}

	return nodes[:count]
}

func findMostRepeatedBytes(data [][]byte) []byte {
	counts := make(map[string]int)

	for _, d := range data {
		counts[string(d)]++
	}

	var mostRepeated []byte
	maxCount := 0

	for d, count := range counts {
		if count > maxCount {
			maxCount = count
			mostRepeated = []byte(d)
		}
	}

	return mostRepeated
}

func findMostRepeatedString(strings []string) string {
	stringCounts := make(map[string]int)

	// Заповнити мапу кількостями повторень, ігноруючи певне значення
	for _, str := range strings {
		if str != "000000000000000000000000000000000000000" {
			stringCounts[str]++
		}
	}

	// Знайти найбільше повторення
	maxCount := 0
	var mostRepeatedString string

	for str, count := range stringCounts {
		if count > maxCount {
			maxCount = count
			mostRepeatedString = str
		}
	}

	return mostRepeatedString
}

func mostRepeatedMap(maps []map[string]string) map[string]string {
	mapCounts := make(map[interface{}]int)

	// Count occurrences of each map in the slice
	for _, m := range maps {
		mapCounts[reflect.ValueOf(m).Pointer()]++
	}

	// Find the map with the maximum count
	maxCount := 0
	var mostRepeatedMap map[string]string

	for _, m := range maps {
		count := mapCounts[reflect.ValueOf(m).Pointer()]
		if count > maxCount {
			maxCount = count
			mostRepeatedMap = m
		}
	}

	return mostRepeatedMap
}
