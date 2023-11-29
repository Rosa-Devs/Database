package db

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Pool struct {
	Database     *Database
	PoolName     string
	Working_path string
}

func (db *Database) GetPool(poolName string) (*Pool, error) {

	folderPath := db.db.DatabasePath + "/" + db.db_name + "/" + poolName // Replace this with the actual path to your folder
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// Folder does not exist
		return nil, err
	}

	return &Pool{Database: db, PoolName: poolName, Working_path: db.db.DatabasePath + "/" + db.db_name + "/" + poolName}, nil
}

func (Pool *Pool) Record(data []byte) error {
	var jsonData map[string]interface{}

	// Unmarshal the JSON data into a map
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return err
	}

	// Generate a random UUID for _id parameter
	id := uuid.New().String()
	jsonData["_id"] = id

	// Convert the updated data back to JSON
	updatedData, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	// Check if the folder with the ID exists, if not, create it
	folderPath := fmt.Sprintf(Pool.Working_path+"/%s", id[:2])
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// Folder doesn't exist, create it
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			fmt.Println("Error creating folder:", err)
		}
	}

	// Save the JSON data to a file
	filePath := fmt.Sprintf(Pool.Working_path+"/"+id[:2]+"/%s.json", id) // Replace this with the actual path where you want to save the file
	err = os.WriteFile(filePath, updatedData, 0644)
	if err != nil {
		return err
	}

	//fmt.Println("New db record!")
	return nil
}

func (p *Pool) GetByID(id string) (map[string]interface{}, error) {
	// Construct the file path using the ID
	filePath := fmt.Sprintf(p.Working_path+"/"+id[:2]+"/%s.json", id) // Replace this with the actual path where you saved the files

	// Read the file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var jsonData map[string]interface{}

	// Unmarshal the JSON data into a map
	if err := json.Unmarshal(fileContent, &jsonData); err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (pool *Pool) Tree() {
	// Read the files in the folder
	files, err := os.ReadDir(pool.Working_path)
	if err != nil {
		fmt.Println("Error reading folder:", err)
		return
	}

	// Print the names of all files in the folder
	fmt.Println("Files in the folder:")
	for _, file := range files {
		if file.IsDir() {
			// Skip directories, print only files
			continue
		}
		fmt.Println(file.Name())
	}
}

func (p *Pool) Filter(filter map[string]interface{}) ([]map[string]interface{}, error) {
	var matchingFiles []string

	// List all files in the directory
	err := p.walkDir(func(filePath string) error {
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Unmarshal the JSON data into a map
		var jsonData map[string]interface{}
		err = json.Unmarshal(fileContent, &jsonData)
		if err != nil {
			return err
		}

		// Check if the file data matches the filter criteria
		matchesFilter := true
		for key, value := range filter {
			// Check if the key exists in the JSON data
			if fieldValue, ok := jsonData[key]; ok {
				// Compare the filter value with the JSON field value (case-insensitive)
				if !strings.EqualFold(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", value)) {
					// If the values don't match, set matchesFilter to false and break the loop
					matchesFilter = false
					break
				}
			}
		}

		// If all filter criteria are met, add the file to the matchingFiles list
		if matchesFilter {
			matchingFiles = append(matchingFiles, jsonData["_id"].(string))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}

	// Iterate through matchingFiles and retrieve data for each ID
	for _, id := range matchingFiles {
		jsonData, err := p.GetByID(id)
		if err != nil {
			log.Println("Error getting matching file:", err)
			continue // Skip to the next file if there's an error
		}

		// Append the unmarshaled JSON data to the data slice
		data = append(data, jsonData)
	}

	return data, nil
}

func (p *Pool) walkDir(callback func(filePath string) error) error {
	return filepath.Walk(p.Working_path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return callback(path)
	})
}

func (p *Pool) Delete(id string) error {
	filePath := fmt.Sprintf(p.Working_path+"/%s.json", id) // Replace this with the actual path where you saved the files

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("record with ID %s not found", id)
	}

	// Delete the file
	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	// fmt.Println("Record with ID", id, "deleted successfully.")
	return nil
}

func (p *Pool) Update(id string, newData map[string]interface{}) error {
	// Get the existing record by ID
	existingData, err := p.GetByID(id)
	if err != nil {
		return err
	}

	// Update the existing record with new data
	for key, value := range newData {
		existingData[key] = value
	}

	// Convert the updated data to JSON
	updatedData, err := json.Marshal(existingData)
	if err != nil {
		return err
	}

	// Write the updated data back to the file
	filePath := fmt.Sprintf(p.Working_path+"/%s.json", id) // Replace this with the actual path where you saved the files
	err = os.WriteFile(filePath, updatedData, 0644)
	if err != nil {
		return err
	}

	//fmt.Println("Record with ID", id, "updated successfully.")
	return nil
}

type LinkS struct {
	Links []Link `json:"links"`
}

// Link struct representing a link
type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// LinkTree function to generate links
func (p *Pool) LinkTree() (LinkS, error) {
	var links []Link

	rootName := "CustomRootName" // Set your custom name here

	err := filepath.Walk(p.Working_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, _ := filepath.Rel(p.Working_path, path)

		// Skip the root directory itself

		fileName := filepath.Base(relativePath)
		dirName := filepath.Base(filepath.Dir(relativePath))

		if dirName == "." {
			return nil
		}
		// Add a link from 2-character folders to the custom root name
		if len(dirName) == 2 {
			links = append(links, Link{Source: dirName, Target: rootName})
		}

		links = append(links, Link{Source: fileName, Target: dirName})

		// Add a link to the folder itself
		if info.IsDir() {
			links = append(links, Link{Source: dirName, Target: filepath.Base(filepath.Dir(relativePath))})
		}

		return nil
	})

	return LinkS{Links: links}, err
}

// func (p *Pool) LinkTree() (LinkS, error) {
// 	var links []Link

// 	rootName := "CustomRootName" // Set your custom name here

// 	err := filepath.Walk(p.Working_path, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		relativePath, _ := filepath.Rel(p.Working_path, path)

// 		// Skip the root directory itself

// 		fileName := filepath.Base(relativePath)
// 		dirName := filepath.Base(filepath.Dir(relativePath))

// 		if dirName == "." {
// 			return nil
// 		}
// 		// Add a link from 2-character folders to the custom root name
// 		if len(dirName) == 2 {
// 			links = append(links, Link{Source: dirName, Target: rootName})
// 		}

// 		links = append(links, Link{Source: fileName, Target: dirName})

// 		// Add a link to the folder itself
// 		if info.IsDir() {
// 			links = append(links, Link{Source: dirName, Target: filepath.Base(filepath.Dir(relativePath))})
// 		}

// 		return nil
// 	})

// 	return LinkS{Links: links}, err
// }
