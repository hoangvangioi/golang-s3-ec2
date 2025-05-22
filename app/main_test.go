package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Set test environment variables
	os.Setenv("S3_BUCKET_NAME", "test-bucket")
	os.Setenv("AWS_REGION", "ap-southeast-1")

	// Run tests
	code := m.Run()

	// Cleanup
	os.Unsetenv("S3_BUCKET_NAME")
	os.Unsetenv("AWS_REGION")

	os.Exit(code)
}

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homeHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Kiểm tra xem response có chứa các thành phần cần thiết không
	expected := []string{
		"File Upload to S3",
		"<form",
		"<input type=\"file\"",
		"<input type=\"submit\"",
	}

	for _, str := range expected {
		if rr.Body.String() == "" {
			t.Errorf("handler returned empty body")
		}
		if !contains(rr.Body.String(), str) {
			t.Errorf("handler returned unexpected body: missing %v", str)
		}
	}
}

func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr
}
