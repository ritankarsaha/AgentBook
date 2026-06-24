package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	MCPPort     string
	AppEnv      string
	FrontendURL string

	DatabaseURL string

	SupabaseURL            string
	SupabaseAnonKey        string
	SupabaseServiceRoleKey string
	SupabaseJWKSURL        string
	SupabaseStorageBucket  string

	NIMAPIKeys      []string
	NIMBaseURL      string
	NIMDefaultModel string

	AgentReplayWebhookSecret string
}

func required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("missing required env var: %s", key))
	}
	return v
}

func Load() *Config {
	nimKeys := parseNIMKeys()

	return &Config{
		Port:        envOr("PORT", "8080"),
		MCPPort:     envOr("MCP_PORT", "8081"),
		AppEnv:      envOr("APP_ENV", "development"),
		FrontendURL: envOr("FRONTEND_URL", "http://localhost:3000"),

		DatabaseURL: required("DATABASE_URL"),

		SupabaseURL:            required("SUPABASE_URL"),
		SupabaseAnonKey:        required("SUPABASE_ANON_KEY"),
		SupabaseServiceRoleKey: required("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseJWKSURL:        required("SUPABASE_JWKS_URL"),
		SupabaseStorageBucket:  envOr("SUPABASE_STORAGE_BUCKET", "agentthreads-media"),

		NIMAPIKeys:      nimKeys,
		NIMBaseURL:      envOr("NVIDIA_NIM_BASE_URL", "https://integrate.api.nvidia.com/v1"),
		NIMDefaultModel: envOr("NVIDIA_NIM_DEFAULT_MODEL", "meta/llama-3.1-70b-instruct"),

		AgentReplayWebhookSecret: os.Getenv("AGENTREPLAY_WEBHOOK_SECRET"),
	}
}

func parseNIMKeys() []string {
	if pool := os.Getenv("NVIDIA_NIM_API_KEYS"); pool != "" {
		var keys []string
		for _, k := range strings.Split(pool, ",") {
			if k = strings.TrimSpace(k); k != "" {
				keys = append(keys, k)
			}
		}
		return keys
	}
	if single := os.Getenv("NVIDIA_NIM_API_KEY"); single != "" {
		return []string{single}
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
