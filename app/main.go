package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
)

const (
	uploadPath  = "./uploads"
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

var (
	s3Client   *s3.S3
	bucketName string
)

func init() {
	// Tạo thư mục uploads nếu chưa tồn tại
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		log.Fatal(err)
	}

	// Khởi tạo AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	}))

	s3Client = s3.New(sess)
	bucketName = os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("S3_BUCKET_NAME environment variable is required")
	}
}

func main() {
	r := mux.NewRouter()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Routes
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/upload", uploadHandler).Methods("POST")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>File Upload</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				max-width: 800px;
				margin: 0 auto;
				padding: 20px;
			}
			.upload-form {
				border: 2px dashed #ccc;
				padding: 20px;
				text-align: center;
				margin: 20px 0;
			}
			.upload-form:hover {
				border-color: #666;
			}
			.file-input {
				margin: 10px 0;
			}
			.submit-btn {
				background-color: #4CAF50;
				color: white;
				padding: 10px 20px;
				border: none;
				border-radius: 4px;
				cursor: pointer;
			}
			.submit-btn:hover {
				background-color: #45a049;
			}
			.result {
				margin-top: 20px;
				padding: 10px;
				border-radius: 4px;
			}
			.success {
				background-color: #dff0d8;
				color: #3c763d;
			}
			.error {
				background-color: #f2dede;
				color: #a94442;
			}
		</style>
	</head>
	<body>
		<h1>File Upload to S3</h1>
		<div class="upload-form">
			<form action="/upload" method="post" enctype="multipart/form-data">
				<input type="file" name="file" class="file-input" required>
				<br>
				<input type="submit" value="Upload" class="submit-btn">
			</form>
		</div>
	</body>
	</html>
	`
	t, err := template.New("home").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Upload to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(header.Filename),
		Body:   file,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate S3 URL
	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, header.Filename)

	// Return success response
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<div class="result success">
			<h3>Upload Successful!</h3>
			<p>File: %s</p>
			<p>S3 URL: <a href="%s" target="_blank">%s</a></p>
		</div>
	`, header.Filename, s3URL, s3URL)
}
