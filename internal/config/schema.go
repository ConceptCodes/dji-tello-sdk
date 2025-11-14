package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// CompileSchema compiles a JSON schema from embedded bytes or fallback filesystem paths.
// schemaID is used as the resource identifier when registering embedded schemas.
func CompileSchema(embedded []byte, fallbackPaths []string, schemaID string) (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	var source string

	if len(embedded) > 0 {
		if err := compiler.AddResource(schemaID, bytes.NewReader(embedded)); err == nil {
			source = schemaID
		}
	}

	if source == "" {
		for _, candidate := range fallbackPaths {
			if candidate == "" {
				continue
			}
			if _, err := os.Stat(candidate); err == nil {
				source = candidate
				break
			}
		}
	}

	if source == "" {
		return nil, fmt.Errorf("schema not found; checked paths: %v", fallbackPaths)
	}

	schema, err := compiler.Compile(source)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema %s: %w", source, err)
	}

	return schema, nil
}

// ValidateJSON validates arbitrary JSON bytes against the compiled schema.
func ValidateJSON(schema *jsonschema.Schema, data []byte) error {
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON for validation: %w", err)
	}

	if err := schema.Validate(instance); err != nil {
		return err
	}

	return nil
}

// ResolvePaths converts any relative paths to absolute form based on the provided base directory.
// This is helpful when callers precompute fallback locations once.
func ResolvePaths(base string, paths ...string) []string {
	var resolved []string
	for _, p := range paths {
		if p == "" {
			continue
		}
		if filepath.IsAbs(p) {
			resolved = append(resolved, p)
			continue
		}
		resolved = append(resolved, filepath.Join(base, p))
	}
	return resolved
}
