package models

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewModelManager(t *testing.T) {
	cacheDir := t.TempDir()

	mm, err := NewModelManager(cacheDir)
	if err != nil {
		t.Fatalf("NewModelManager failed: %v", err)
	}

	if mm == nil {
		t.Fatal("ModelManager should not be nil")
	}

	// Verify cache directory was created
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Fatal("Cache directory should exist")
	}

	// Verify models are initialized from registry
	if len(mm.models) == 0 {
		t.Fatal("ModelManager should have models from registry")
	}
}

func TestGetModelPath_Cached(t *testing.T) {
	cacheDir := t.TempDir()

	// Use a registry model name and create the cached file
	modelName := "yolo-v8n"
	modelPath := filepath.Join(cacheDir, modelName)
	if err := os.WriteFile(modelPath, []byte("fake model data"), 0644); err != nil {
		t.Fatalf("Failed to create test model file: %v", err)
	}

	// Create manager to pick up the cached model
	mm, err := NewModelManager(cacheDir)
	if err != nil {
		t.Fatalf("NewModelManager failed: %v", err)
	}

	// Get path for the cached model
	path, err := mm.GetModelPath(modelName)
	if err != nil {
		t.Fatalf("GetModelPath failed for cached model: %v", err)
	}

	if path != modelPath {
		t.Errorf("Expected path %s, got %s", modelPath, path)
	}
}

func TestGetModelPath_NotCached(t *testing.T) {
	cacheDir := t.TempDir()

	// Create a mock HTTP server to serve model
	modelContent := []byte("fake model content for download")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(modelContent)))
		w.WriteHeader(http.StatusOK)
		io.Copy(w, io.NopCloser(io.LimitReader(io.NopCloser(io.MultiReader()), int64(len(modelContent)))))
		w.Write(modelContent)
	}))
	defer server.Close()

	mm, err := NewModelManager(cacheDir)
	if err != nil {
		t.Fatalf("NewModelManager failed: %v", err)
	}

	// Create a custom model info pointing to our test server
	modelInfo := &ModelInfo{
		Name:     "test-download-model",
		URL:      server.URL + "/model.onnx",
		Size:     int64(len(modelContent)),
		Checksum: "",
		Format:   "onnx",
	}
	mm.models[modelInfo.Name] = modelInfo

	// Download the model
	progressChan := make(chan *DownloadProgress, 1)
	err = mm.DownloadModel(modelInfo, progressChan)
	if err != nil {
		t.Fatalf("DownloadModel failed: %v", err)
	}

	// Wait for download to complete
	progress := <-progressChan
	if progress.Error != "" {
		t.Fatalf("Download failed: %s", progress.Error)
	}

	if !progress.Completed {
		t.Fatal("Download should be completed")
	}

	// Now get the path - should return the downloaded model
	path, err := mm.GetModelPath(modelInfo.Name)
	if err != nil {
		t.Fatalf("GetModelPath failed after download: %v", err)
	}

	expectedPath := filepath.Join(cacheDir, modelInfo.Name)
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Downloaded model file should exist")
	}
}
