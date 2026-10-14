package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Requirement struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

func getRequirementsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	requirements := []Requirement{
		{ID: 1, Text: "The submit button must remain disabled until a valid screenshot is pasted."},
		{ID: 2, Text: "The preview image container must display a dashed border when empty."},
		{ID: 3, Text: "All uploaded images must automatically generate a timestamped filename."},
		{ID: 4, Text: "Backend router must properly support CORS preflight OPTIONS requests."},
	}

	json.NewEncoder(w).Encode(requirements)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Base() strips any directory parts so a crafted filename cannot escape uploadDir.
	filename := filepath.Base(handler.Filename)
	if filename == "." || filename == string(filepath.Separator) {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file data", http.StatusInternalServerError)
		return
	}

	// Save Draft / Save send a status; store the pins next to the image as <image>.json.
	if status := r.FormValue("status"); status != "" {
		if status != "draft" && status != "final" {
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}
		var pins []json.RawMessage
		if err := json.Unmarshal([]byte(r.FormValue("pins")), &pins); err != nil {
			http.Error(w, "Invalid pins", http.StatusBadRequest)
			return
		}
		mapping, _ := json.MarshalIndent(map[string]any{
			"image":  filename,
			"status": status,
			"pins":   pins,
		}, "", "  ")
		jsonPath := filepath.Join(uploadDir, strings.TrimSuffix(filename, filepath.Ext(filename))+".json")
		if err := os.WriteFile(jsonPath, mapping, 0o644); err != nil {
			http.Error(w, "Unable to save mapping", http.StatusInternalServerError)
			return
		}
		fmt.Printf("Saved %s mapping: %s\n", status, jsonPath)
	}

	fmt.Printf("Successfully uploaded: %s\n", filename)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Upload successful"))
}

func main() {
	http.HandleFunc("/api/upload", uploadHandler)
	http.HandleFunc("/api/requirements", getRequirementsHandler)

	fmt.Println("Backend server starting on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
