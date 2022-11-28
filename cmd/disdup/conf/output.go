package clconf

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	config "github.com/ethanv2/disdup/conf"
	"github.com/ethanv2/disdup/output"
)

// An Output is a json-encodable representation of a disdup output.
type Output struct {
	Type      string                 `json:"type"`
	Arguments map[string]interface{} `json:"args"`
}

// convertOutput converts a temporary representation of an output to the format
// which can be read by disdup.
func convertOutput(name string, tmpl Output, cfg *config.Config) error {
	var out output.Output

	switch tmpl.Type {
	case "stdout":
		out = &output.Writer{
			Output:  os.Stdout,
			Collate: 0,
		}
	}

	cfg.Outputs = append(cfg.Outputs, config.OutputConfig{Name: name, Output: out})
	return nil
}

func loadOutputs(cfg *config.Config, paths []string) error {
	found := false
	var outputs map[string]Output
	for _, dir := range paths {
		path := filepath.Join(dir, OutputConfigName)
		f, err := os.Open(path)
		if err == nil {
			defer f.Close()

			found = true
			buf, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("outputs config: %w", ErrIO)
			}

			buf = processConfig(buf)
			err = json.Unmarshal(buf, &outputs)
			if err != nil {
				return fmt.Errorf("outputs config: bad syntax: %w", err)
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("outputs config: %w", ErrNotFound)
	}

	for name, output := range outputs {
		if err := convertOutput(name, output, cfg); err != nil {
			return fmt.Errorf("outputs config: output %s (%s): %w", name, output.Type, ErrOutput)
		}
	}

	return nil
}
