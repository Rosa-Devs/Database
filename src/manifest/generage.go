package manifest

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"
)

func GenerateNoise(length int) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(randomString)
}

func GenereateManifest(name string, file bool, opts string) *Manifest {

	// Create pubsub indentifier for database

	pubsub := GenerateNoise(64)
	chiper := GenerateNoise(32)

	model := Manifest{
		Name:     name,
		PubSub:   pubsub,
		Optional: opts,
		Chiper:   chiper,
		UId:      GenerateNoise(64),
	}

	if file {
		// Convert the data structure to JSON
		jsonData, err := json.MarshalIndent(model, "", "    ")
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			return nil
		}

		// Save the JSON to a file
		err = os.WriteFile(name+".json", jsonData, 0644)
		if err != nil {
			log.Println("Error writing to file:", err)
			return nil
		}

		log.Println("Create new Manifest " + name + ".json")
		log.Println("Give it your friend to connect to the database!")
		return nil
	}

	return &model

}
