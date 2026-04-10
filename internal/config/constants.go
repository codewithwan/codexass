package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	ClientID          = "app_EMoamEEZ73f0CkXaXp7hrann"
	AuthEndpoint      = "https://auth.openai.com/oauth/authorize"
	TokenEndpoint     = "https://auth.openai.com/oauth/token"
	UsageURL          = "https://chatgpt.com/backend-api/wham/usage"
	CodexResponsesURL = "https://chatgpt.com/backend-api/codex/responses"
	CallbackHost      = "127.0.0.1"
	CallbackPort      = 1455
	CallbackPath      = "/auth/callback"
	DefaultModel      = "gpt-5.3-codex"
	DefaultOriginator = "codexass"
	StoreDirName      = "codexass"
	StoreDirFlag      = "--store-dir"
)

func DefaultStoreDir() string {
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, StoreDirName)
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "." + StoreDirName
	}
	return filepath.Join(home, "."+StoreDirName)
}
