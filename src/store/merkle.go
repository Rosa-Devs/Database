package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// generateHash generates a SHA-256 hash for the given data.
func generateHash(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// generateMerkleTree generates a Merkle tree for the given directory.
// calculateHashTree calculates the hash tree for the given directory.
func (p *Database) GenereateMerkleTree() (string, error) {
	rootPath := p.db.DatabasePath + "/" + p.manifest.UId
	var fileList []string

	// Traverse the directory and get the list of file paths
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileList = append(fileList, path)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Sort the file paths for consistency
	sort.Strings(fileList)

	// Generate leaf nodes for each file and calculate hashes
	hashTree := make(map[string]string)
	for _, filePath := range fileList {
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		hashTree[filePath] = generateHash(fileData)
	}

	// Build the Merkle tree bottom-up
	for len(hashTree) > 1 {
		var nextLevel = make(map[string]string)

		// Combine pairs of hashes to create parent hashes
		keys := make([]string, 0, len(hashTree))
		for key := range hashTree {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for i := 0; i < len(keys); i += 2 {
			hash1 := hashTree[keys[i]]
			var hash2 string

			if i+1 < len(keys) {
				hash2 = hashTree[keys[i+1]]
			}

			parentHash := generateHash([]byte(hash1 + hash2))
			nextLevel[keys[i]] = parentHash
		}

		hashTree = nextLevel
	}

	if len(hashTree) == 0 {
		return "000000000000000000000000000000000000000", nil
	}
	rootHash := hashTree[fileList[0]]

	return rootHash, nil
}

func (db *Database) Index() (map[string]string, error) {

	rootPath := db.db.DatabasePath + "/" + db.manifest.UId
	var fileList []string

	// Traverse the directory and get the list of file paths
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileList = append(fileList, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort the file paths for consistency
	sort.Strings(fileList)

	// Generate leaf nodes for each file and calculate hashes
	hashIndex := make(map[string]string)
	for _, filePath := range fileList {
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		index := strings.Index(filePath, "/")

		// If "/" is found, trim the path
		var trimmedPath string
		if index != -1 {
			trimmedPath = filePath[index+1:]
			//fmt.Println(trimmedPath)
		} else {
			fmt.Println("Path does not contain '/'")
		}

		hashIndex[trimmedPath] = generateHash(fileData)
	}

	return hashIndex, nil
}

func (db *Database) CalculateChangedFiles(prevHashTree, currentHashTree map[string]string) []string {
	var changedFiles []string

	// Compare the hashes in the current hash tree with the previous hash tree
	for filePath, currentHash := range currentHashTree {
		prevHash, exists := prevHashTree[filePath]
		if !exists || prevHash != currentHash {
			// File has changed or is new
			changedFiles = append(changedFiles, filePath)
		}
	}

	// Check for deleted files
	for filePath := range prevHashTree {
		_, exists := currentHashTree[filePath]
		if !exists {
			// File has been deleted
			changedFiles = append(changedFiles, filePath)
		}
	}

	return changedFiles
}
