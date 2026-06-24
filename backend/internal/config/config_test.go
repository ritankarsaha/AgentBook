package config

import "testing"

func setAllRequired(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://fake:fake@localhost:5432/fake")
	t.Setenv("SUPABASE_URL", "https://fake.supabase.co")
	t.Setenv("SUPABASE_ANON_KEY", "fake-anon-key")
	t.Setenv("SUPABASE_SERVICE_ROLE_KEY", "fake-service-role-key")
	t.Setenv("SUPABASE_JWKS_URL", "https://fake.supabase.co/auth/v1/.well-known/jwks.json")
}

func TestLoad_AllRequiredPresent_DoesNotPanic(t *testing.T) {
	setAllRequired(t)

	cfg := Load()

	if cfg.DatabaseURL == "" {
		t.Fatal("expected DatabaseURL to be loaded")
	}
}

func TestLoad_MissingRequiredVar_Panics(t *testing.T) {
	required := []string{
		"DATABASE_URL",
		"SUPABASE_URL",
		"SUPABASE_ANON_KEY",
		"SUPABASE_SERVICE_ROLE_KEY",
		"SUPABASE_JWKS_URL",
	}

	for _, missing := range required {
		t.Run(missing, func(t *testing.T) {
			setAllRequired(t)
			t.Setenv(missing, "")

			defer func() {
				if recover() == nil {
					t.Fatalf("expected Load() to panic when %s is missing, it did not", missing)
				}
			}()
			Load()
		})
	}
}
