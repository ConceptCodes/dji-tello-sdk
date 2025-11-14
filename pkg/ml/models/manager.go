package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// modelRegistry contains predefined model information
var modelRegistry = map[string]*ModelInfo{
	"yolo-v8n": {
		Name:         "yolo-v8n",
		Version:      "8.0.0",
		URL:          "file://./models/yolov8n_real.onnx", // Local test model for now
		Size:         1020,                                // ~1KB test model
		Checksum:     "",                                  // Will be calculated after download
		Description:  "YOLOv8 Nano - Smallest and fastest model for real-time applications",
		Format:       "onnx",
		Architecture: "cnn",
	},
	"yolo-v8s": {
		Name:         "yolo-v8s",
		Version:      "8.3.0",
		URL:          "https://github.com/ultralytics/ultralytics/releases/download/v8.3.0/yolov8s.onnx",
		Size:         21889684, // ~21MB
		Checksum:     "",       // Will be calculated after download
		Description:  "YOLOv8 Small - Balanced model for good accuracy and speed",
		Format:       "onnx",
		Architecture: "cnn",
	},
	"yolo-v8m": {
		Name:         "yolo-v8m",
		Version:      "8.3.0",
		URL:          "https://github.com/ultralytics/ultralytics/releases/download/v8.3.0/yolov8m.onnx",
		Size:         49747008, // ~47MB
		Checksum:     "",       // Will be calculated after download
		Description:  "YOLOv8 Medium - High accuracy model for most applications",
		Format:       "onnx",
		Architecture: "cnn",
	},
	"yolo-v8l": {
		Name:         "yolo-v8l",
		Version:      "8.3.0",
		URL:          "https://github.com/ultralytics/ultralytics/releases/download/v8.3.0/yolov8l.onnx",
		Size:         83696625, // ~80MB
		Checksum:     "",       // Will be calculated after download
		Description:  "YOLOv8 Large - Highest accuracy model for offline processing",
		Format:       "onnx",
		Architecture: "cnn",
	},
}

// ModelInfo contains metadata about a model
type ModelInfo struct {
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	URL          string    `json:"url"`
	Size         int64     `json:"size"`
	Checksum     string    `json:"checksum"`
	Description  string    `json:"description"`
	Format       string    `json:"format"`
	Architecture string    `json:"architecture"`
	DownloadedAt time.Time `json:"downloaded_at,omitempty"`
	FilePath     string    `json:"file_path,omitempty"`
}

// ModelManager handles model downloading, caching, and management
type ModelManager struct {
	cacheDir   string
	httpClient *http.Client
	models     map[string]*ModelInfo
	downloads  map[string]*DownloadProgress
	logger     *logrus.Logger
}

// DownloadProgress tracks download progress
type DownloadProgress struct {
	ModelName  string    `json:"model_name"`
	TotalBytes int64     `json:"total_bytes"`
	Downloaded int64     `json:"downloaded"`
	Speed      int64     `json:"speed"` // bytes per second
	StartTime  time.Time `json:"start_time"`
	ETA        time.Time `json:"eta,omitempty"`
	Percentage float64   `json:"percentage"`
	Completed  bool      `json:"completed"`
	Error      string    `json:"error,omitempty"`
}

// NewModelManager creates a new model manager
func NewModelManager(cacheDir string) (*ModelManager, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	mm := &ModelManager{
		cacheDir: cacheDir,
		httpClient: &http.Client{
			Timeout: 30 * time.Minute,
		},
		models:    make(map[string]*ModelInfo),
		downloads: make(map[string]*DownloadProgress),
		logger:    logrus.New(),
	}

	// Initialize model registry
	mm.initializeModelRegistry()

	// Load existing models from cache
	if err := mm.loadCachedModels(); err != nil {
		mm.logger.Warnf("Failed to load cached models: %v", err)
	}

	return mm, nil
}

// initializeModelRegistry initializes the model registry
func (mm *ModelManager) initializeModelRegistry() {
	// Copy predefined models to the models map
	for name, modelInfo := range modelRegistry {
		// Create a copy to avoid modifying the registry
		modelCopy := *modelInfo
		mm.models[name] = &modelCopy
	}

	mm.logger.Infof("Initialized model registry with %d models", len(modelRegistry))
}

// DownloadModel downloads a model from URL
func (mm *ModelManager) DownloadModel(modelInfo *ModelInfo, progressChan chan<- *DownloadProgress) error {
	if _, exists := mm.downloads[modelInfo.Name]; exists {
		return fmt.Errorf("download already in progress for model: %s", modelInfo.Name)
	}

	// Check if model already exists
	if mm.IsModelDownloaded(modelInfo.Name) {
		mm.logger.Infof("Model %s already exists", modelInfo.Name)
		return nil
	}

	// Initialize download progress
	progress := &DownloadProgress{
		ModelName:  modelInfo.Name,
		TotalBytes: modelInfo.Size,
		StartTime:  time.Now(),
		Completed:  false,
	}
	mm.downloads[modelInfo.Name] = progress

	// Start download in goroutine
	go mm.downloadModelAsync(modelInfo, progress, progressChan)

	return nil
}

// downloadModelAsync performs the actual download
func (mm *ModelManager) downloadModelAsync(modelInfo *ModelInfo, progress *DownloadProgress, progressChan chan<- *DownloadProgress) {
	defer func() {
		delete(mm.downloads, modelInfo.Name)
		if progressChan != nil {
			progressChan <- progress
		}
	}()

	// Create temporary file
	tempPath := filepath.Join(mm.cacheDir, modelInfo.Name+".tmp")
	file, err := os.Create(tempPath)
	if err != nil {
		progress.Error = fmt.Sprintf("Failed to create temp file: %v", err)
		return
	}
	defer file.Close()

	// Start HTTP request
	resp, err := mm.httpClient.Get(modelInfo.URL)
	if err != nil {
		progress.Error = fmt.Sprintf("Failed to start download: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		progress.Error = fmt.Sprintf("HTTP error: %s", resp.Status)
		return
	}

	// Create progress reader
	reader := &progressReader{
		reader:   resp.Body,
		progress: progress,
		logger:   mm.logger,
	}

	// Download with progress tracking
	hash := sha256.New()
	multiWriter := io.MultiWriter(file, hash)

	bytesWritten, err := io.Copy(multiWriter, reader)
	if err != nil {
		progress.Error = fmt.Sprintf("Download failed: %v", err)
		os.Remove(tempPath)
		return
	}

	// Verify checksum (if provided)
	calculatedChecksum := hex.EncodeToString(hash.Sum(nil))
	if modelInfo.Checksum != "" && calculatedChecksum != modelInfo.Checksum {
		progress.Error = fmt.Sprintf("Checksum mismatch: expected %s, got %s", modelInfo.Checksum, calculatedChecksum)
		os.Remove(tempPath)
		return
	}

	// Verify size
	if bytesWritten != modelInfo.Size {
		progress.Error = fmt.Sprintf("Size mismatch: expected %d, got %d", modelInfo.Size, bytesWritten)
		os.Remove(tempPath)
		return
	}

	// Move temp file to final location
	finalPath := filepath.Join(mm.cacheDir, modelInfo.Name)
	if err := os.Rename(tempPath, finalPath); err != nil {
		progress.Error = fmt.Sprintf("Failed to save model: %v", err)
		os.Remove(tempPath)
		return
	}

	// Update model info
	modelInfo.DownloadedAt = time.Now()
	modelInfo.FilePath = finalPath
	mm.models[modelInfo.Name] = modelInfo

	// Mark as completed
	progress.Completed = true
	progress.Percentage = 100.0
	mm.logger.Infof("Successfully downloaded model: %s", modelInfo.Name)
}

// IsModelDownloaded checks if a model is already downloaded
func (mm *ModelManager) IsModelDownloaded(modelName string) bool {
	modelInfo, exists := mm.models[modelName]
	if !exists {
		return false
	}

	if modelInfo.FilePath == "" {
		return false
	}

	// Check if file still exists
	if _, err := os.Stat(modelInfo.FilePath); err != nil {
		return false
	}

	return true
}

// GetModel returns information about a model
func (mm *ModelManager) GetModel(modelName string) (*ModelInfo, error) {
	// First check if model exists in our map
	modelInfo, exists := mm.models[modelName]
	if !exists {
		// Check if it's in the registry but not loaded yet
		registryModel, registryExists := modelRegistry[modelName]
		if !registryExists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}

		// Add to our models map
		modelCopy := *registryModel
		mm.models[modelName] = &modelCopy
		return &modelCopy, nil
	}

	return modelInfo, nil
}

// GetModelPath returns the file path for a model, resolving model names to paths
func (mm *ModelManager) GetModelPath(modelName string) (string, error) {
	// If it looks like a file path, return as-is
	if filepath.Ext(modelName) != "" {
		// Check if file exists
		if _, err := os.Stat(modelName); err != nil {
			return "", fmt.Errorf("model file not found: %s", modelName)
		}
		return modelName, nil
	}

	// Treat as model name and resolve to downloaded path
	modelInfo, err := mm.GetModel(modelName)
	if err != nil {
		return "", err
	}

	if modelInfo.FilePath != "" {
		// Check if file still exists
		if _, err := os.Stat(modelInfo.FilePath); err != nil {
			return "", fmt.Errorf("model file not found: %s", modelInfo.FilePath)
		}
		return modelInfo.FilePath, nil
	}

	// Check if model file exists in cache directory with .onnx extension
	cachePath := filepath.Join(mm.cacheDir, modelName+".onnx")
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	// Check if model file exists in cache directory with exact name
	cachePath = filepath.Join(mm.cacheDir, modelName)
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	return "", fmt.Errorf("model not downloaded: %s", modelName)
}

// ListModels returns all available models (including registry models)
func (mm *ModelManager) ListModels() []*ModelInfo {
	var models []*ModelInfo

	// Add all models from registry
	for name, registryModel := range modelRegistry {
		modelCopy := *registryModel

		// Check if model is downloaded and update info
		if downloadedModel, exists := mm.models[name]; exists && downloadedModel.FilePath != "" {
			modelCopy.FilePath = downloadedModel.FilePath
			modelCopy.DownloadedAt = downloadedModel.DownloadedAt
		}

		models = append(models, &modelCopy)
	}

	return models
}

// ListDownloadedModels returns only downloaded models
func (mm *ModelManager) ListDownloadedModels() []*ModelInfo {
	var models []*ModelInfo
	for _, model := range mm.models {
		if mm.IsModelDownloaded(model.Name) {
			models = append(models, model)
		}
	}
	return models
}

// ListRegistryModels returns all models from registry (for CLI help)
func (mm *ModelManager) ListRegistryModels() []*ModelInfo {
	var models []*ModelInfo
	for _, model := range modelRegistry {
		modelCopy := *model
		models = append(models, &modelCopy)
	}
	return models
}

// RemoveModel removes a model from cache
func (mm *ModelManager) RemoveModel(modelName string) error {
	modelInfo, exists := mm.models[modelName]
	if !exists {
		return fmt.Errorf("model not found: %s", modelName)
	}

	// Remove file if it exists
	if modelInfo.FilePath != "" {
		if err := os.Remove(modelInfo.FilePath); err != nil {
			mm.logger.Warnf("Failed to remove model file: %v", err)
		}
	}

	// Remove from registry
	delete(mm.models, modelName)

	mm.logger.Infof("Removed model: %s", modelName)
	return nil
}

// ValidateModel validates a downloaded model
func (mm *ModelManager) ValidateModel(modelName string) error {
	modelInfo, exists := mm.models[modelName]
	if !exists {
		return fmt.Errorf("model not found: %s", modelName)
	}

	if modelInfo.FilePath == "" {
		return fmt.Errorf("model file not available: %s", modelName)
	}

	// Check file exists
	if _, err := os.Stat(modelInfo.FilePath); err != nil {
		return fmt.Errorf("model file not found: %w", err)
	}

	// Verify checksum
	file, err := os.Open(modelInfo.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open model file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	calculatedChecksum := hex.EncodeToString(hash.Sum(nil))
	if calculatedChecksum != modelInfo.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", modelInfo.Checksum, calculatedChecksum)
	}

	// Verify size
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if stat.Size() != modelInfo.Size {
		return fmt.Errorf("size mismatch: expected %d, got %d", modelInfo.Size, stat.Size())
	}

	return nil
}

// GetDownloadProgress returns current download progress
func (mm *ModelManager) GetDownloadProgress(modelName string) (*DownloadProgress, bool) {
	progress, exists := mm.downloads[modelName]
	return progress, exists
}

// CancelDownload cancels an ongoing download
func (mm *ModelManager) CancelDownload(modelName string) error {
	progress, exists := mm.downloads[modelName]
	if !exists {
		return fmt.Errorf("no download in progress for model: %s", modelName)
	}

	progress.Error = "Download cancelled"
	delete(mm.downloads, modelName)

	mm.logger.Infof("Cancelled download for model: %s", modelName)
	return nil
}

// loadCachedModels loads existing models from cache directory
func (mm *ModelManager) loadCachedModels() error {
	// This would load from a metadata file
	// For now, we'll implement basic scanning
	entries, err := os.ReadDir(mm.cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip temporary files
		if filepath.Ext(entry.Name()) == ".tmp" {
			continue
		}

		// Create basic model info for existing files
		filePath := filepath.Join(mm.cacheDir, entry.Name())
		stat, err := entry.Info()
		if err != nil {
			continue
		}

		modelInfo := &ModelInfo{
			Name:         entry.Name(),
			FilePath:     filePath,
			Size:         stat.Size(),
			DownloadedAt: stat.ModTime(),
		}

		mm.models[entry.Name()] = modelInfo
	}

	return nil
}

// progressReader wraps an io.Reader to track progress
type progressReader struct {
	reader    io.Reader
	progress  *DownloadProgress
	logger    *logrus.Logger
	lastTime  time.Time
	lastBytes int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)

	if n > 0 {
		pr.progress.Downloaded += int64(n)
		pr.progress.Percentage = float64(pr.progress.Downloaded) / float64(pr.progress.TotalBytes) * 100.0

		// Calculate speed
		now := time.Now()
		if !pr.lastTime.IsZero() {
			elapsed := now.Sub(pr.lastTime).Seconds()
			if elapsed > 0 {
				bytesDiff := pr.progress.Downloaded - pr.lastBytes
				pr.progress.Speed = int64(float64(bytesDiff) / elapsed)

				// Calculate ETA
				if pr.progress.Speed > 0 {
					remaining := pr.progress.TotalBytes - pr.progress.Downloaded
					secondsRemaining := float64(remaining) / float64(pr.progress.Speed)
					pr.progress.ETA = now.Add(time.Duration(secondsRemaining) * time.Second)
				}
			}
		}

		pr.lastTime = now
		pr.lastBytes = pr.progress.Downloaded
	}

	return n, err
}
