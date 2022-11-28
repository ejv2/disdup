// Package clconf loads CLient Conf from configuration files.
package clconf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	config "github.com/ethanv2/disdup/conf"
)

// Key os-independent file names or paths.
const (
	PrimaryConfigName = "disdup.conf"
	OutputConfigName  = "outputs.conf"
)

// Generic config loading related errors.
var (
	ErrNotFound = errors.New("not found")
	ErrSyntax   = errors.New("bad syntax")
	ErrIO       = errors.New("I/O error")
	ErrOutput   = errors.New("invalid output")
)

func globalConfigDir() string {
	if runtime.GOOS == "windows" {
		return "C:\\ProgramData\\disdup"
	}

	return "/etc/disdup"
}

// processConfig preprocesses the config json to allow comments.
func processConfig(buf []byte) []byte {
	newbuf := make([]byte, 0, len(buf))
	for _, elem := range strings.Split(string(buf), "\n") {
		if !strings.HasPrefix(elem, "//") {
			newbuf = append(newbuf, []byte(elem)...)
		}
	}

	return newbuf
}

func loadPrimary(cfg *config.Config, paths []string) error {
	found := false
	for _, path := range paths {
		f, err := os.Open(filepath.Join(path, PrimaryConfigName))
		if err == nil {
			defer f.Close()

			found = true
			in, err := io.ReadAll(f)
			in = processConfig(in)
			if err != nil {
				return fmt.Errorf("primary config: %w", ErrIO)
			}

			err = json.Unmarshal([]byte(in), cfg)
			if err != nil {
				return fmt.Errorf("primary config: bad syntax: %w", err)
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("primary config: %w", ErrNotFound)
	}

	return nil
}

func LoadConfig() (config.Config, error) {
	cfg := config.Config{}

	cfgpath, _ := os.UserConfigDir()
	paths := []string{
		".",
		cfgpath,
		globalConfigDir(),
	}

	err := loadPrimary(&cfg, paths)
	if err != nil {
		return config.Config{}, err
	}

	err = loadOutputs(&cfg, paths)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}
