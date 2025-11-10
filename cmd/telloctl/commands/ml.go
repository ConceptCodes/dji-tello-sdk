package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/config"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/models"
	"github.com/spf13/cobra"
)

var (
	configDir string
	verbose   bool
)

// mlCmd represents the ML command
var mlCmd = &cobra.Command{
	Use:   "ml",
	Short: "Machine Learning commands for Tello SDK",
	Long: `Manage ML processors, configurations, and pipelines for the Tello SDK.
	
This command group provides tools for:
- Managing ML configurations
- Validating configuration files
- Listing available processors
- Creating default configurations`,
}

// mlInitCmd represents the ml init command
var mlInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ML configuration",
	Long: `Create default ML configuration files and schemas.
	
This command will create:
- Default ML pipeline configuration
- Default processor configurations
- JSON schemas for validation`,
	RunE: runMLInit,
}

// mlValidateCmd represents the ml validate command
var mlValidateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate ML configuration",
	Long: `Validate an ML configuration file against JSON schemas.
	
This command checks:
- JSON syntax validity
- Schema compliance
- Processor configuration validity
- Required field presence`,
	Args: cobra.ExactArgs(1),
	RunE: runMLValidate,
}

// mlListCmd represents the ml list command
var mlListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available configurations",
	Long:  `List all available ML configuration files in the config directory.`,
	RunE:  runMLList,
}

// mlProcessorsCmd represents the ml processors command
var mlProcessorsCmd = &cobra.Command{
	Use:   "processors",
	Short: "List available ML processors",
	Long:  `List all registered ML processor types and their capabilities.`,
	RunE:  runMLProcessors,
}

// mlConfigCmd represents the ml config command
var mlConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Manage ML configurations including creation, validation, and listing.`,
}

// mlModelsCmd represents the ml models command
var mlModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Model management commands",
	Long:  `Manage ML models including download, list, and cleanup operations.`,
}

// mlModelsListCmd represents the ml models list command
var mlModelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List downloaded models",
	Long:  `List all downloaded ML models with their status and metadata.`,
	RunE:  runMLModelsList,
}

// mlModelsDownloadCmd represents the ml models download command
var mlModelsDownloadCmd = &cobra.Command{
	Use:   "download [model-name]",
	Short: "Download a model",
	Long: `Download a specific ML model from the repository.
	
Available models:
- yolo-v8n: YOLOv8 Nano (smallest, fastest)
- yolo-v8s: YOLOv8 Small (balanced)
- yolo-v8m: YOLOv8 Medium (accurate)
- yolo-v8l: YOLOv8 Large (most accurate)`,
	Args: cobra.ExactArgs(1),
	RunE: runMLModelsDownload,
}

// mlModelsInfoCmd represents the ml models info command
var mlModelsInfoCmd = &cobra.Command{
	Use:   "info [model-name]",
	Short: "Show model information",
	Long:  `Display detailed information about a specific model.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMLModelsInfo,
}

// mlModelsCleanupCmd represents the ml models cleanup command
var mlModelsCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up unused models",
	Long:  `Remove unused or corrupted model files to free up disk space.`,
	RunE:  runMLModelsCleanup,
}

// addMLCommands adds ML commands to the root command
func addMLCommands() {
	// Add flags to ML command
	mlCmd.PersistentFlags().StringVar(&configDir, "config-dir", "configs", "Configuration directory")
	mlCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Add subcommands
	mlCmd.AddCommand(mlInitCmd)
	mlCmd.AddCommand(mlValidateCmd)
	mlCmd.AddCommand(mlListCmd)
	mlCmd.AddCommand(mlProcessorsCmd)
	mlCmd.AddCommand(mlConfigCmd)
	mlCmd.AddCommand(mlModelsCmd)

	// Add config subcommands
	mlConfigCmd.AddCommand(mlInitCmd)
	mlConfigCmd.AddCommand(mlValidateCmd)
	mlConfigCmd.AddCommand(mlListCmd)

	// Add models subcommands
	mlModelsCmd.AddCommand(mlModelsListCmd)
	mlModelsCmd.AddCommand(mlModelsDownloadCmd)
	mlModelsCmd.AddCommand(mlModelsInfoCmd)
	mlModelsCmd.AddCommand(mlModelsCleanupCmd)
}

// MLCmd returns the ML command for external registration
func MLCmd() *cobra.Command {
	addMLCommands()
	return mlCmd
}

// runMLInit initializes ML configuration
func runMLInit(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Printf("Initializing ML configuration in directory: %s\n", configDir)
	}

	// Create config manager
	configManager := config.NewConfigManager(configDir)

	// Create schemas directory and default schema files first
	if err := createSchemaFiles(configDir); err != nil {
		return fmt.Errorf("failed to create schema files: %w", err)
	}

	// Load schemas
	schemas := []struct {
		name string
		file string
	}{
		{"ml-pipeline", "ml-pipeline-schema.json"},
		{"yolo", "yolo-config-schema.json"},
	}

	for _, schema := range schemas {
		if err := configManager.LoadSchema(schema.name, schema.file); err != nil {
			fmt.Printf("Warning: Failed to load schema %s: %v\n", schema.file, err)
		}
	}

	// Create default configurations
	if err := configManager.CreateDefaultConfigs(); err != nil {
		return fmt.Errorf("failed to create default configurations: %w", err)
	}

	fmt.Printf("✅ ML configuration initialized successfully!\n")
	fmt.Printf("📁 Configuration directory: %s\n", configDir)
	fmt.Printf("📄 Created files:\n")

	// List created files
	files, err := os.ReadDir(configDir)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
				fmt.Printf("   - %s\n", file.Name())
			}
		}
	}

	schemaDir := filepath.Join(configDir, "schemas")
	if _, err := os.Stat(schemaDir); err == nil {
		fmt.Printf("📋 Schema files:\n")
		schemaFiles, err := os.ReadDir(schemaDir)
		if err == nil {
			for _, file := range schemaFiles {
				if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
					fmt.Printf("   - %s\n", file.Name())
				}
			}
		}
	}

	return nil
}

// runMLValidate validates ML configuration
func runMLValidate(cmd *cobra.Command, args []string) error {
	configFile := args[0]

	if verbose {
		fmt.Printf("Validating configuration file: %s\n", configFile)
	}

	// Create config manager
	configManager := config.NewConfigManager(configDir)

	// Load schemas first
	schemas := []string{
		"ml-pipeline-schema.json",
		"yolo-config-schema.json",
	}

	for _, schema := range schemas {
		schemaPath := filepath.Join(configDir, "schemas", schema)
		if _, err := os.Stat(schemaPath); err == nil {
			schemaName := "ml-pipeline"
			if schema == "yolo-config-schema.json" {
				schemaName = "yolo"
			}

			if err := configManager.LoadSchema(schemaName, schema); err != nil {
				fmt.Printf("Warning: Failed to load schema %s: %v\n", schema, err)
			}
		}
	}

	// Check if file exists
	if !configManager.ConfigExists(configFile) {
		return fmt.Errorf("configuration file not found: %s", configFile)
	}

	// Validate configuration
	result, err := configManager.ValidateConfigFromFile(configFile, "ml-pipeline")
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if result.Valid() {
		fmt.Printf("✅ Configuration file %s is valid!\n", configFile)
	} else {
		fmt.Printf("❌ Configuration file %s has validation errors:\n", configFile)
		for _, err := range result.Errors() {
			fmt.Printf("   - %s\n", err)
		}
		return fmt.Errorf("configuration validation failed")
	}

	return nil
}

// runMLList lists available configurations
func runMLList(cmd *cobra.Command, args []string) error {
	configManager := config.NewConfigManager(configDir)

	configs, err := configManager.ListConfigs()
	if err != nil {
		return fmt.Errorf("failed to list configurations: %w", err)
	}

	if len(configs) == 0 {
		fmt.Printf("No configuration files found in %s\n", configDir)
		fmt.Printf("Run 'telloctl ml init' to create default configurations.\n")
		return nil
	}

	fmt.Printf("📄 Available ML configurations in %s:\n", configDir)
	for _, config := range configs {
		fmt.Printf("   - %s\n", config)
	}

	return nil
}

// runMLProcessors lists available processors
func runMLProcessors(cmd *cobra.Command, args []string) error {
	fmt.Printf("🤖 Available ML processors:\n")

	// This would be populated from the processor registry
	// For now, we'll show the known types
	processors := []struct {
		name        string
		description string
		status      string
	}{
		{"yolo", "YOLO object detection", "Available"},
		{"face", "Face detection and recognition", "Available"},
		{"slam", "Simultaneous Localization and Mapping", "Planned"},
		{"gesture", "Gesture recognition", "Planned"},
		{"segmentation", "3D segmentation", "Planned"},
		{"custom", "Custom processor", "Available"},
	}

	for _, proc := range processors {
		statusIcon := "✅"
		if proc.status == "Planned" {
			statusIcon = "🚧"
		}

		fmt.Printf("   %s %s - %s (%s)\n", statusIcon, proc.name, proc.description, proc.status)
	}

	fmt.Printf("\n💡 Use 'telloctl ml init' to create processor configurations.\n")

	return nil
}

// runMLModelsList lists downloaded models
func runMLModelsList(cmd *cobra.Command, args []string) error {
	modelManager, err := models.NewModelManager("models")
	if err != nil {
		return fmt.Errorf("failed to create model manager: %w", err)
	}

	modelList := modelManager.ListDownloadedModels()

	if len(modelList) == 0 {
		fmt.Printf("No models found. Use 'telloctl ml models download <model-name>' to download models.\n")
		return nil
	}

	fmt.Printf("📦 Downloaded ML models:\n\n")
	for _, model := range modelList {
		fmt.Printf("🤖 %s\n", model.Name)
		fmt.Printf("   Size: %s\n", formatBytes(model.Size))
		fmt.Printf("   Path: %s\n", model.FilePath)
		if model.Description != "" {
			fmt.Printf("   Description: %s\n", model.Description)
		}
		if !model.DownloadedAt.IsZero() {
			fmt.Printf("   Downloaded: %s\n", model.DownloadedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}

	return nil
}

// runMLModelsDownload downloads a model
func runMLModelsDownload(cmd *cobra.Command, args []string) error {
	modelName := args[0]

	modelManager, err := models.NewModelManager("models")
	if err != nil {
		return fmt.Errorf("failed to create model manager: %w", err)
	}

	// Check if model already exists
	if modelManager.IsModelDownloaded(modelName) {
		fmt.Printf("Model '%s' already exists. Use 'telloctl ml models info %s' for details.\n", modelName, modelName)
		return nil
	}

	// Get model info for requested model
	modelInfo, err := modelManager.GetModel(modelName)
	if err != nil {
		// Show available models if not found
		fmt.Printf("❌ Model '%s' not found.\n\n", modelName)
		fmt.Printf("📦 Available models:\n")
		registryModels := modelManager.ListRegistryModels()
		for _, model := range registryModels {
			status := "📥 Available"
			if modelManager.IsModelDownloaded(model.Name) {
				status = "✅ Downloaded"
			}
			fmt.Printf("   %s %s - %s (%s)\n", status, model.Name, model.Description, formatBytes(model.Size))
		}
		return fmt.Errorf("model '%s' not found", modelName)
	}

	fmt.Printf("📥 Downloading model: %s\n", modelName)

	// Download model with progress callback
	progressChan := make(chan *models.DownloadProgress, 100)

	// Start progress monitor
	done := make(chan bool)
	go func() {
		for progress := range progressChan {
			if progress.Error != "" {
				fmt.Printf("\r❌ Error: %s\n", progress.Error)
				done <- true
				return
			}
			if progress.Completed {
				fmt.Printf("\r✅ Download completed!\n")
				done <- true
				return
			}
			fmt.Printf("\r⏳ Downloading... %.1f%%", progress.Percentage*100)
		}
	}()

	err = modelManager.DownloadModel(modelInfo, progressChan)
	if err != nil {
		close(progressChan)
		return fmt.Errorf("failed to start download: %w", err)
	}

	// Wait for download to complete
	<-done
	close(progressChan)

	// Get updated model info
	downloadedModel, err := modelManager.GetModel(modelName)
	if err != nil {
		return fmt.Errorf("failed to get updated model info: %w", err)
	}

	fmt.Printf("✅ Model '%s' downloaded successfully!\n", downloadedModel.Name)
	fmt.Printf("📁 Path: %s\n", downloadedModel.FilePath)
	fmt.Printf("📏 Size: %s\n", formatBytes(downloadedModel.Size))

	return nil
}

// runMLModelsInfo shows model information
func runMLModelsInfo(cmd *cobra.Command, args []string) error {
	modelName := args[0]

	modelManager, err := models.NewModelManager("models")
	if err != nil {
		return fmt.Errorf("failed to create model manager: %w", err)
	}

	model, err := modelManager.GetModel(modelName)
	if err != nil {
		return fmt.Errorf("failed to get model info: %w", err)
	}

	fmt.Printf("🤖 Model Information: %s\n", model.Name)
	fmt.Printf("   Status: %s\n", func() string {
		if modelManager.IsModelDownloaded(modelName) {
			return "✅ Downloaded"
		}
		return "📥 Not Downloaded"
	}())
	fmt.Printf("   Size: %s\n", formatBytes(model.Size))
	if model.FilePath != "" {
		fmt.Printf("   Path: %s\n", model.FilePath)
	}
	fmt.Printf("   Description: %s\n", model.Description)
	if !model.DownloadedAt.IsZero() {
		fmt.Printf("   Downloaded: %s\n", model.DownloadedAt.Format("2006-01-02 15:04:05"))
	}

	if model.Checksum != "" {
		fmt.Printf("   SHA256: %s\n", model.Checksum)
	}

	return nil
}

// runMLModelsCleanup cleans up unused models
func runMLModelsCleanup(cmd *cobra.Command, args []string) error {
	modelManager, err := models.NewModelManager("models")
	if err != nil {
		return fmt.Errorf("failed to create model manager: %w", err)
	}

	fmt.Printf("🧹 Scanning for corrupted models...\n")

	// Get all downloaded models
	modelList := modelManager.ListDownloadedModels()

	var toRemove []*models.ModelInfo
	for _, model := range modelList {
		// Validate model
		if err := modelManager.ValidateModel(model.Name); err != nil {
			fmt.Printf("❌ Model %s is corrupted: %v\n", model.Name, err)
			toRemove = append(toRemove, model)
		}
	}

	if len(toRemove) == 0 {
		fmt.Printf("✅ All models are valid. No cleanup needed.\n")
		return nil
	}

	fmt.Printf("Found %d corrupted model(s):\n", len(toRemove))
	for _, model := range toRemove {
		fmt.Printf("   - %s (%s)\n", model.Name, formatBytes(model.Size))
	}

	fmt.Printf("\nRemove these models? [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		for _, model := range toRemove {
			if err := modelManager.RemoveModel(model.Name); err != nil {
				fmt.Printf("❌ Failed to remove %s: %v\n", model.Name, err)
			} else {
				fmt.Printf("✅ Removed %s\n", model.Name)
			}
		}
		fmt.Printf("🧹 Cleanup completed. Freed %s\n", formatBytes(calculateTotalSizePtr(toRemove)))
	} else {
		fmt.Printf("Cleanup cancelled.\n")
	}

	return nil
}

// formatBytes formats bytes in human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// calculateTotalSize calculates total size of models
func calculateTotalSize(modelList []models.ModelInfo) int64 {
	var total int64
	for _, model := range modelList {
		total += model.Size
	}
	return total
}

// calculateTotalSizePtr calculates total size of model pointers
func calculateTotalSizePtr(modelList []*models.ModelInfo) int64 {
	var total int64
	for _, model := range modelList {
		total += model.Size
	}
	return total
}

// createSchemaFiles creates the schema files in the configs/schemas directory
func createSchemaFiles(configDir string) error {
	schemaDir := filepath.Join(configDir, "schemas")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		return fmt.Errorf("failed to create schemas directory: %w", err)
	}

	// ML Pipeline Schema
	mlPipelineSchema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "processors": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "type": {"type": "string", "enum": ["yolo", "face", "slam", "gesture", "segmentation", "custom"]},
          "enabled": {"type": "boolean"},
          "priority": {"type": "integer", "minimum": 0},
          "config_file": {"type": "string"}
        },
        "required": ["name", "type", "enabled", "priority"]
      }
    },
    "pipeline": {
      "type": "object",
      "properties": {
        "max_concurrent_processors": {"type": "integer", "minimum": 1},
        "frame_buffer_size": {"type": "integer", "minimum": 1},
        "worker_pool_size": {"type": "integer", "minimum": 1},
        "enable_metrics": {"type": "boolean"},
        "target_fps": {"type": "integer", "minimum": 1, "maximum": 60}
      },
      "required": ["max_concurrent_processors", "frame_buffer_size", "worker_pool_size", "enable_metrics", "target_fps"]
    },
    "overlay": {
      "type": "object",
      "properties": {
        "enabled": {"type": "boolean"},
        "show_fps": {"type": "boolean"},
        "show_detections": {"type": "boolean"},
        "show_tracking": {"type": "boolean"},
        "show_confidence": {"type": "boolean"},
        "colors": {"type": "object"},
        "line_width": {"type": "integer", "minimum": 1},
        "font_size": {"type": "integer", "minimum": 1},
        "font_scale": {"type": "number", "minimum": 0.1}
      },
      "required": ["enabled", "show_fps", "show_detections", "show_tracking", "show_confidence"]
    }
  },
  "required": ["processors", "pipeline", "overlay"]
}`

	// YOLO Config Schema
	yoloSchema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "model": {"type": "string"},
    "confidence": {"type": "number", "minimum": 0, "maximum": 1},
    "nms_threshold": {"type": "number", "minimum": 0, "maximum": 1},
    "input_size": {
      "type": "array",
      "items": {"type": "integer"},
      "minItems": 2,
      "maxItems": 2
    },
    "classes": {
      "type": "array",
      "items": {"type": "string"}
    }
  },
  "required": ["model", "confidence", "nms_threshold", "input_size", "classes"]
}`

	// Write schema files
	schemaFiles := map[string]string{
		"ml-pipeline-schema.json": mlPipelineSchema,
		"yolo-config-schema.json": yoloSchema,
	}

	for filename, content := range schemaFiles {
		schemaPath := filepath.Join(schemaDir, filename)
		if err := os.WriteFile(schemaPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write schema file %s: %w", filename, err)
		}
	}

	return nil
}
