package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{name: "defaults when missing", want: "full"},
		{name: "overrides values", data: `{"server":{"port":9090},"cache_ttl":"5m","default_mode":"path_only"}`, want: "path_only"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if tt.data != "" && os.WriteFile(path, []byte(tt.data), 0o600) != nil {
				t.Fatal("write config")
			}
			cfg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.DefaultMode != tt.want {
				t.Fatalf("mode = %q, want %q", cfg.DefaultMode, tt.want)
			}
		})
	}
}
