package main

import (
	_ "embed"
	"fmt"
	"gocrypt/pkg/cryptolib"
	"net/http"
	"os"
)

//go:embed index.html
var indexHTML []byte

func runServe(port string) {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/encrypt", handleEncrypt)
	http.HandleFunc("/decrypt", handleDecrypt)

	fmt.Printf("Server listening on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Server failed: %v\n", err)
		os.Exit(1)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write(indexHTML)
}

func handleEncrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 10MB limit for demo purposes (can be increased)
	// ParseMultipartForm parses a request body as multipart/form-data.
	// The whole request body is parsed and up to a total of maxMemory bytes of
	// its file parts are stored in memory, with the remainder stored on
	// disk in temporary files.
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	password := r.FormValue("password")
	if password == "" {
		http.Error(w, "Password required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.enc\"", header.Filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Encrypt directly to ResponseWriter
	err = cryptolib.EncryptStream(file, w, password, header.Size, nil)
	if err != nil {
		// If streaming has started, we can't really change the status code cleanly, 
		// but we can log it.
		fmt.Printf("Encryption error: %v\n", err)
	}
}

func handleDecrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	password := r.FormValue("password")
	if password == "" {
		http.Error(w, "Password required", http.StatusBadRequest)
		return
	}

	// Remove .enc extension if present
    outFilename := header.Filename
    if len(outFilename) > 4 && outFilename[len(outFilename)-4:] == ".enc" {
        outFilename = outFilename[:len(outFilename)-4]
    } else {
        outFilename = outFilename + ".dec"
    }

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", outFilename))
	w.Header().Set("Content-Type", "application/octet-stream")

	err = cryptolib.DecryptStream(file, w, password, header.Size, nil)
	if err != nil {
		fmt.Printf("Decryption error: %v\n", err)
	}
}
