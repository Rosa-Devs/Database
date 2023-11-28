package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	db "github.com/Rosa-Devs/POC/src/store"
	"github.com/google/uuid"
)

func main() {

	// !! GLOBAl DB MANAGER !!
	//CREATE DATABSE INSTANCE
	Drvier := db.DB{}
	//START DATABSE INSTANCE
	Drvier.Start("test_db_1")
	//CREATE TEST DB
	Drvier.CreateDb("test")

	// !! WORKING WITH SPECIFIED BATABASE !!
	db := Drvier.GetDb("test")

	err := db.CreatePool("test_pool")
	if err != nil {
		log.Println("Mayby this pool alredy exist:", err)
		//return
	}

	pool, err := db.GetPool("test_pool")
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		//SIMULATE ADDING DATA
		rand.Seed(time.Now().UnixNano())
		for {
			// Generate random data
			randomData := map[string]interface{}{
				"field1": rand.Intn(100),             // Random integer between 0 and 100
				"field2": rand.Float64() * 100,       // Random float between 0 and 100
				"field3": uuid.New().String(),        // Random UUID as a string
				"field4": time.Now().UnixNano(),      // Current timestamp in nanoseconds
				"field5": fmt.Sprintf("Record%d", 1), // Custom string with record number
			}

			// Convert data to JSON
			jsonData, err := json.Marshal(randomData)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			// Call Record function to save the record
			err = pool.Record(jsonData)
			if err != nil {
				fmt.Println("Error recording data:", err)
				return
			}
			time.Sleep(time.Millisecond * 50)
		}
	}()

	go func() {
		for {
			filter := map[string]interface{}{
				"field1": 96, // Random integer between 0 and 100
			}

			data, err := pool.Filter(filter)
			if err != nil {
				fmt.Println("Data:", data)
				fmt.Println("Error filtering data:", err)
			}
			log.Println(data)
			time.Sleep(time.Millisecond * 70)
		}
	}()

	for {
	}

}
