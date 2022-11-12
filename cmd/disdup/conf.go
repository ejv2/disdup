package main

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
)

// Generic config loading related errors.
var (
	ErrNotFound = errors.New("not found")
	ErrSyntax   = errors.New("bad syntax")
	ErrIO       = errors.New("I/O error")
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

func LoadConfig() (config.Config, error) {
	cfg := config.Config{}

	cfgpath, _ := os.UserConfigDir()
	paths := [...]string{
		".",
		cfgpath,
		globalConfigDir(),
	}

	found := false
	for _, path := range paths {
		f, err := os.Open(filepath.Join(path, PrimaryConfigName))
		if err == nil {
			defer f.Close()

			found = true
			in, err := io.ReadAll(f)
			in = processConfig(in)
			if err != nil {
				return config.Config{}, ErrIO
			}

			err = json.Unmarshal([]byte(in), &cfg)
			if err != nil {
				return config.Config{}, fmt.Errorf("bad syntax: %w", err)
			}
			break
		}
	}

	if !found {
		return config.Config{}, ErrNotFound
	}

	return cfg, nil
}
