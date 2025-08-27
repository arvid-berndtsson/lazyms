package config

import (
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"
)

func TestLoadMissingReturnsDefaults(t *testing.T) {
	d := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", d)
	t.Cleanup(func(){ os.Unsetenv("XDG_CONFIG_HOME") })

	cfg, err := Load()
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if cfg.PreferredAuth == "" { t.Fatalf("expected default preferredAuth=cli or non-empty") }
}

func TestLoadParsesYAML(t *testing.T) {
	d := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", d)
	t.Cleanup(func(){ os.Unsetenv("XDG_CONFIG_HOME") })

	path := filepath.Join(d, "lazyms", "config.yaml")
	os.MkdirAll(filepath.Dir(path), 0o755)
	data := []byte("tenantId: 11111111-1111-1111-1111-111111111111\nclientId: abc\npreferredAuth: devicecode\n")
	if err := os.WriteFile(path, data, 0o644); err != nil { t.Fatal(err) }

	cfg, err := Load()
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if cfg.TenantID == "" || cfg.ClientID == "" || cfg.PreferredAuth != "devicecode" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

// guard against accidental use of iotest import
var _ = iotest.ErrTimeout


