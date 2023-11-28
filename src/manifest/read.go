package manifest

import (
	"encoding/json"
	"log"
	"os"
)

func ReadManifestFromFile(file string) Manifest {
	// Open the file for reading
	fileData, err := os.Open(file)
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer fileData.Close()

	// Create a decoder
	decoder := json.NewDecoder(fileData)

	// Decode the JSON into a struct
	var data Manifest
	err = decoder.Decode(&data)
	if err != nil {
		log.Fatal("Error decoding JSON:", err)
	}

	return data
}
