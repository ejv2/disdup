package clconf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	config "github.com/ethanv2/disdup/conf"
	"github.com/ethanv2/disdup/output"
)

// Output parsing or processing errors.
var (
	ErrWrongType      = errors.New("unexpected type")
	ErrUnknownCollate = errors.New("unknown collation mode")
)

// An Output is a json-encodable representation of a disdup output.
type Output struct {
	Type      string                 `json:"type"`
	Arguments map[string]interface{} `json:"args"`
}

func parseCollation(conf map[string]interface{}) (int, error) {
	if rcoll, ok := conf["collate"]; ok {
		// Unwrap option into map of options
		coll, ok := rcoll.(string)
		if !ok {
			return 0, fmt.Errorf("key collate: %w", ErrWrongType)
		}

		switch coll {
		case "channel":
			return output.WriterCollateChannel, nil
		case "user":
			return output.WriterCollateUser, nil
		default:
			return 0, fmt.Errorf("%s: %w", coll, ErrUnknownCollate)
		}
	}

	return 0, nil
}

// convertOutput converts a temporary representation of an output to the format
// which can be read by disdup.
func convertOutput(name string, tmpl Output, cfg *config.Config) error {
	var out output.Output

	switch tmpl.Type {
	case "stdout":
		col, err := parseCollation(tmpl.Arguments)
		if err != nil {
			return fmt.Errorf("output %s: %w", name, err)
		}
		rprefix, ok := tmpl.Arguments["prefix"]
		prefix := ""
		if ok {
			if prefix, ok = rprefix.(string); !ok {
				return fmt.Errorf("output %s: key prefix: %w", name, ErrWrongType)
			}
			prefix += " " // Append a space to properly space output in log format
		}
		out = &output.Writer{
			Output:  os.Stdout,
			Collate: col,
			Prefix:  prefix,
		}
	default:
		return ErrOutput
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
			return fmt.Errorf("outputs config: output %s (%s): %w", name, output.Type, err)
		}
	}

	return nil
}
