package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadWithoutFileOrEnvUsesDefaults(t *testing.T) {
	t.Setenv("CONFIG_PATH", filepath.Join(t.TempDir(), "missing.yaml"))
	clearLLMEnv(t)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.LLM.Timeout != DefaultLLMTimeout {
		t.Fatalf("expected default llm timeout %s, got %s", DefaultLLMTimeout, cfg.LLM.Timeout)
	}
	cfg.LLM.Timeout = 0
	if cfg != (Config{}) {
		t.Fatalf("unexpected non-default config fields: %+v", cfg)
	}
}

func TestLoadFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	yaml := `
httpAddr: ":9090"
llm:
  provider: openai
  timeout: 5s
  baseUrl: https://example.test
`
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	clearLLMEnv(t)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("addr: %q", cfg.HTTPAddr)
	}
	if cfg.LLM.Provider != "openai" || cfg.LLM.Timeout != 5*time.Second {
		t.Fatalf("llm: %+v", cfg.LLM)
	}
}

func TestEnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("llm:\n  provider: openai\n  timeout: 5s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LLM_PROVIDER", "mock")
	t.Setenv("LLM_TIMEOUT", "10s")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.LLM.Provider != "mock" || cfg.LLM.Timeout != 10*time.Second {
		t.Fatalf("env did not override: %+v", cfg.LLM)
	}
}

func clearLLMEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"LLM_PROVIDER", "LLM_TIMEOUT", "LLM_BASE_URL", "LLM_API_KEY", "HTTP_ADDR"} {
		t.Setenv(k, "")
		_ = os.Unsetenv(k)
	}
}
