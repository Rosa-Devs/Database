package db

import (
	"crypto/sha256"
	"encoding/hex"
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
func (p *Pool) GenereateMerkleTree() (string, map[string]string, error) {
	rootPath := p.Working_path
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
		return "", nil, err
	}

	// Sort the file paths for consistency
	sort.Strings(fileList)

	// Generate leaf nodes for each file and calculate hashes
	hashTree := make(map[string]string)
	for _, filePath := range fileList {
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return "", nil, err
		}

		hashTree[filePath] = generateHash(fileData)
	}

	// Save tree
	tree := hashTree

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
		return "000000000000000000000000000000000000000", nil, nil
	}
	rootHash := hashTree[fileList[0]]

	return rootHash, tree, nil
}

// func (p *Pool) CalculateChangedFiles(prevHashTree, currentHashTree map[string]string) []string {
// 	var changedFiles []string

// 	// Compare the hashes in the current hash tree with the previous hash tree
// 	for filePath, currentHash := range currentHashTree {
// 		// Normalize file path to lowercase for case-insensitive comparison
// 		normalizedFilePath := strings.ToLower(filePath)

// 		prevHash, exists := prevHashTree[normalizedFilePath]
// 		if !exists || prevHash != currentHash {
// 			// File has changed or is new
// 			changedFiles = append(changedFiles, filePath)
// 		}
// 	}

// 	// Check for deleted files
// 	for filePath := range prevHashTree {
// 		// Normalize file path to lowercase for case-insensitive comparison
// 		normalizedFilePath := strings.ToLower(filePath)

// 		_, exists := currentHashTree[normalizedFilePath]
// 		if !exists {
// 			// File has been deleted
// 			changedFiles = append(changedFiles, filePath)
// 		}
// 	}

// 	return changedFiles
// }

func (p *Pool) CalculateChangedFiles(prevHashTree, currentHashTree map[string]string) []string {
	var changedFiles []string

	// Compare the hashes in the current hash tree with the previous hash tree
	for filePath, currentHash := range currentHashTree {
		// Normalize file path to lowercase for case-insensitive comparison
		normalizedFilePath := strings.ToLower(filePath)

		// Extract only the file name without path and extension
		fileName := strings.TrimSuffix(filepath.Base(normalizedFilePath), filepath.Ext(normalizedFilePath))

		prevHash, exists := prevHashTree[normalizedFilePath]
		if !exists || prevHash != currentHash {
			// File has changed or is new
			changedFiles = append(changedFiles, fileName)
		}
	}

	// Check for deleted files
	for filePath := range prevHashTree {
		// Normalize file path to lowercase for case-insensitive comparison
		normalizedFilePath := strings.ToLower(filePath)

		// Extract only the file name without path and extension
		fileName := strings.TrimSuffix(filepath.Base(normalizedFilePath), filepath.Ext(normalizedFilePath))

		_, exists := currentHashTree[normalizedFilePath]
		if !exists {
			// File has been deleted
			changedFiles = append(changedFiles, fileName)
		}
	}

	return changedFiles
}
