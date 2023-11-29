package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	db "github.com/Rosa-Devs/POC/src/store"
)

//go:embed static/*
var content embed.FS

func main() {
	port := flag.Int("port", 8080, "Port to run the HTTP server on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Path[1:] // Remove leading "/"
		if filePath == "" {
			filePath = "static/index.html"
		} else {
			filePath = "static/" + filePath
		}

		file, err := content.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// Determine content type based on file extension
		contentType := http.DetectContentType(file)
		w.Header().Set("Content-Type", contentType)
		w.Write(file)
	})

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

	http.HandleFunc("/tree", func(w http.ResponseWriter, r *http.Request) {
		fileTreeHandler(w, r, pool)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Server is running on http://localhost%s\n", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func fileTreeHandler(w http.ResponseWriter, r *http.Request, p *db.Pool) {
	links, err := p.LinkTree()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.MarshalIndent(links, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
