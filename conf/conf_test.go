package config_test

import (
	"testing"

	config "github.com/ejv2/disdup/conf"
	"github.com/ejv2/disdup/output"
)

func testConfigUseNorm(t *testing.T) {
	cfg := &config.Config{}
	cfg.Use("writer", &output.Writer{}).
		Use("chan", &output.Channel{}).
		Use("rawchan", &output.RawChannel{})
}

func testConfigUseDup(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("Duplicate output registration did not panic")
		}
	}()

	cfg := &config.Config{}
	cfg.Use("first", nil).
		Use("second", nil).
		Use("second", nil)
}

func testConfigGuild(t *testing.T) {
}

func TestConfig(t *testing.T) {
	t.Run("Use", func(t *testing.T) {
		t.Run("Norm", testConfigUseNorm)
		t.Run("Duplicate", testConfigUseDup)
	})

	t.Run("Guild", func(t *testing.T) {
		t.Run("Norm", testConfigGuild)
	})
}
