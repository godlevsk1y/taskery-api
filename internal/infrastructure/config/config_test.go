package config_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/config"
	"github.com/stretchr/testify/require"
)

func TestMustLoad_OK(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("function unexpectedly panicked: %v", r)
		}
	}()

	envFile, err := os.Create("./.env")
	require.NoError(t, err)
	defer os.Remove(".env")
	defer envFile.Close()

	n, err := envFile.WriteString("CONFIG_PATH=test.yml")
	require.NoError(t, err)
	require.NotZero(t, n)

	configFile, err := os.Create("test.yml")
	require.NoError(t, err)
	defer os.Remove("test.yml")
	defer configFile.Close()

	configText := strings.Replace(`
		env: "local" # local, dev, production

		http_server:
		  address: "localhost:6666"
		  timeout: 15s
		  idle_timeout: 90s
		
		postgres_connection:
		  host: "localhost"
		  port: "0000"
		  username: "name"
		  password: "pswrd"
		  db_name: "db_name"`, "\t", "", -1)

	n, err = configFile.Write([]byte(configText))
	require.NoError(t, err)
	require.NotZero(t, n)

	cfg := config.MustLoad()
	require.Equal(t, "local", cfg.Environment)
	require.Equal(t, "localhost:6666", cfg.HTTPServer.Address)
	require.Equal(t, 15*time.Second, cfg.HTTPServer.Timeout)
	require.Equal(t, 90*time.Second, cfg.HTTPServer.IdleTimeout)
}

func TestMustLoad_NoEnv(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("function did not panic as expected")
		}
	}()

	_ = config.MustLoad()
}

func TestMustLoad_NoConfigPath(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("function did not panic as expected")
		}
	}()

	envFile, err := os.Create("./.env")
	require.NoError(t, err)
	defer os.Remove(".env")
	defer envFile.Close()

	_ = config.MustLoad()
}

func TestMustLoad_ConfigFileNotExists(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("function did not panic as expected")
		}
	}()

	envFile, err := os.Create("./.env")
	require.NoError(t, err)
	defer os.Remove(".env")
	defer envFile.Close()

	n, err := envFile.WriteString("CONFIG_PATH=test.yml")
	require.NoError(t, err)
	require.NotZero(t, n)

	_ = config.MustLoad()
}

func TestMustLoad_ConfigParsingFailed(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("function did not panic as expected")
		}
	}()

	envFile, err := os.Create("./.env")
	require.NoError(t, err)
	defer os.Remove(".env")
	defer envFile.Close()

	n, err := envFile.WriteString("CONFIG_PATH=test.yml")
	require.NoError(t, err)
	require.NotZero(t, n)

	configFile, err := os.Create("test.yml")
	require.NoError(t, err)
	defer os.Remove("test.yml")
	defer configFile.Close()

	n, err = configFile.WriteString("wklenfk4ho24t08ur0[hf[o2'h")

	_ = config.MustLoad()
}
