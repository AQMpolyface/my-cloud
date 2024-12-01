package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	database "server/packages"
	"strings"
)

const uploadDir = "./uploads"

func Initialize() error {
	return os.MkdirAll(uploadDir, os.ModePerm)
}

// CheckAuth verifies if the user is authenticated via UUID cookie
func CheckAuth(w http.ResponseWriter, r *http.Request, db *sql.DB) bool {
	cookie, err := r.Cookie("uuid")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return false
		}
		http.Error(w, "Error retrieving cookie", http.StatusInternalServerError)
		return false
	}

	valid, err := CheckUuid(db, cookie.Value)
	if err != nil {
		fmt.Println("error retrieving uuid from db:", err)
		return false
	}

	if !valid {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return false
	}

	return true
}

// UploadHandler handles file uploads
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := database.ConnectToDB()
	if err != nil {
		http.Error(w, "somthing went wrong with the upload", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	if !CheckAuth(w, r, db) {
		http.Error(w, "you arent authorized bozo", http.StatusUnauthorized)
		return
	}

	// Parse multipart form with 10MB max memory
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		fmt.Println("error parsing file", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("error formatting file", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create file in uploads directory
	dst, err := os.Create(filepath.Join(uploadDir, handler.Filename))
	if err != nil {
		fmt.Println("error creating file", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, file); err != nil {
		fmt.Println("error copying:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
}

// DownloadHandler handles file downloads
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	db, err := database.ConnectToDB()
	if err != nil {
		http.Error(w, "somthing went wrong with downloading", http.StatusInternalServerError)
		return
	}

	defer db.Close()
	if !CheckAuth(w, r, db) {
		http.Error(w, "you arent authorized bozo", http.StatusUnauthorized)
		return
	}

	// Extract filename from path
	filename := strings.TrimPrefix(r.URL.Path, "/api/download/")
	if filename == "" {
		http.Error(w, "No filename specified", http.StatusBadRequest)
		return
	}

	uploadDir, err := filepath.Abs("uploads")
	if err != nil {
		log.Fatalf("Failed to determine absolute path for uploads directory: %v", err)
	}

	// Clean the filename to prevent directory traversal
	filename = filepath.Clean(filename)
	filePath := filepath.Join(uploadDir, filename)

	// Ensure the file is within the uploads directory
	if !strings.HasPrefix(filePath, uploadDir) {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("file not found:", filePath)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(w, r, filePath)
}

// ListFilesHandler handles listing of available files
func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	db, err := database.ConnectToDB()
	if err != nil {
		http.Error(w, "somthing went wrong when fetching files", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	if !CheckAuth(w, r, db) {
		http.Error(w, "you arent authorized bozo", http.StatusUnauthorized)
		return
	}

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "[\n")
	for i, file := range files {
		if i > 0 {
			fmt.Fprint(w, ",\n")
		}
		fmt.Fprintf(w, `  {"name": "%s"}`, file.Name())
	}
	fmt.Fprint(w, "\n]")
}

func CheckUuid(db *sql.DB, uuid string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM authentification WHERE UUID = ?", uuid).Scan(&count)
	if err != nil {
		return false, err
	}

	// If count is greater than 0, the UUID exists so true is retuwurned
	return count > 0, nil
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
func main() {
	if err := Initialize(); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	http.HandleFunc("/api/upload", UploadHandler)
	http.HandleFunc("/api/download/", DownloadHandler)
	http.HandleFunc("/api/files", ListFilesHandler)
	http.HandleFunc("/", IndexHandler)

	fmt.Println("Server started at :8088")
	log.Fatal(http.ListenAndServe(":8088", nil))
}
