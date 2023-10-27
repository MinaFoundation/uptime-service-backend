package delegation_backend

import (
	"fmt"
	"os"
	"testing"
)

type MockLogger struct {
	lastMessage string
}

// Debug implements log.EventLogger.
func (*MockLogger) Debug(args ...interface{}) {
	panic("unimplemented")
}

// Debugf implements log.EventLogger.
func (*MockLogger) Debugf(format string, args ...interface{}) {
	panic("unimplemented")
}

// Error implements log.EventLogger.
func (*MockLogger) Error(args ...interface{}) {
	panic("unimplemented")
}

// Errorf implements log.EventLogger.
func (*MockLogger) Errorf(format string, args ...interface{}) {
	panic("unimplemented")
}

// Info implements log.EventLogger.
func (*MockLogger) Info(args ...interface{}) {
	panic("unimplemented")
}

// Infof implements log.EventLogger.
func (*MockLogger) Infof(format string, args ...interface{}) {
	panic("unimplemented")
}

// Panic implements log.EventLogger.
func (*MockLogger) Panic(args ...interface{}) {
	panic("unimplemented")
}

// Panicf implements log.EventLogger.
func (*MockLogger) Panicf(format string, args ...interface{}) {
	panic("unimplemented")
}

// Warn implements log.EventLogger.
func (*MockLogger) Warn(args ...interface{}) {
	panic("unimplemented")
}

// Warnf implements log.EventLogger.
func (*MockLogger) Warnf(format string, args ...interface{}) {
	panic("unimplemented")
}

func (m *MockLogger) Fatal(args ...interface{}) {
	m.lastMessage = fmt.Sprint(args...)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.lastMessage = fmt.Sprintf(format, args...)
}

func TestLoadEnv(t *testing.T) {
	mockLogger := &MockLogger{}

	t.Run("load AWS from config file", func(t *testing.T) {
		// Create a temporary config file
		fileContent := `
			{
				"network_name": "test_network",
				"gsheet_id": "test_gsheet_id",
				"delegation_whitelist_list": "test_list",
				"delegation_whitelist_column": "test_column",
				"aws": {
					"account_id": "test_account_id",
					"bucket_name_suffix": "test_suffix",
					"region": "test_region",
					"access_key_id": "test_access_key_id",
					"secret_access_key": "test_secret_access_key"
				}
			}
			`
		tmpFile := "/tmp/test_config.json"
		os.WriteFile(tmpFile, []byte(fileContent), 0644)
		os.Setenv("CONFIG_FILE", tmpFile)
		config := LoadEnv(mockLogger)
		if config.NetworkName != "test_network" {
			t.Errorf("Expected network_name to be test_network but got %s", config.NetworkName)
		}
		if config.Aws == nil {
			t.Errorf("Expected Aws config to load but got %s", config.Aws)
		}
		os.Unsetenv("CONFIG_FILE")
	})

	t.Run("load DB from config file", func(t *testing.T) {
		// Create a temporary config file
		fileContent := `
			{
				"network_name": "test_network",
				"gsheet_id": "test_gsheet_id",
				"delegation_whitelist_list": "test_list",
				"delegation_whitelist_column": "test_column",
				"database": {
					"connection_string": "test_connection_string",
					"type": "test_type"
				}
			}
			`
		tmpFile := "/tmp/test_config.json"
		os.WriteFile(tmpFile, []byte(fileContent), 0644)
		os.Setenv("CONFIG_FILE", tmpFile)
		config := LoadEnv(mockLogger)
		if config.NetworkName != "test_network" {
			t.Errorf("Expected network_name to be test_network but got %s", config.NetworkName)
		}
		if config.Database == nil {
			t.Errorf("Expected Database config to load but got %s", config.Aws)
		}
		os.Unsetenv("CONFIG_FILE")
	})

	t.Run("load Filesystem from config file", func(t *testing.T) {
		// Create a temporary config file
		fileContent := `
			{
				"network_name": "test_network",
				"gsheet_id": "test_gsheet_id",
				"delegation_whitelist_list": "test_list",
				"delegation_whitelist_column": "test_column",
				"filesystem": {
					"path": "test_path"
				}
			}
			`
		tmpFile := "/tmp/test_config.json"
		os.WriteFile(tmpFile, []byte(fileContent), 0644)
		os.Setenv("CONFIG_FILE", tmpFile)
		config := LoadEnv(mockLogger)
		if config.NetworkName != "test_network" {
			t.Errorf("Expected network_name to be test_network but got %s", config.NetworkName)
		}
		if config.LocalFileSystem == nil {
			t.Errorf("Expected LocalFileSystem config to load but got %s", config.Aws)
		}
		os.Unsetenv("CONFIG_FILE")
	})

	t.Run("load AWS from env", func(t *testing.T) {
		os.Setenv("CONFIG_NETWORK_NAME", "test_network")
		os.Setenv("CONFIG_GSHEET_ID", "test_gsheet_id")
		os.Setenv("DELEGATION_WHITELIST_LIST", "test_list")
		os.Setenv("DELEGATION_WHITELIST_COLUMN", "test_column")
		os.Setenv("AWS_ACCESS_KEY_ID", "test_access_key_id")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test_secret_access_key")
		os.Setenv("CONFIG_AWS_REGION", "test_region")
		os.Setenv("CONFIG_AWS_ACCOUNT_ID", "test_account_id")
		os.Setenv("CONFIG_BUCKET_NAME_SUFFIX", "test_suffix")

		config := LoadEnv(mockLogger)
		if config.Aws == nil || config.Aws.AccessKeyId != "test_access_key_id" {
			t.Error("Failed to load AWS configs from environment variables")
		}

		// Cleanup
		os.Unsetenv("CONFIG_NETWORK_NAME")
		os.Unsetenv("CONFIG_GSHEET_ID")
		os.Unsetenv("DELEGATION_WHITELIST_LIST")
		os.Unsetenv("DELEGATION_WHITELIST_COLUMN")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("CONFIG_AWS_REGION")
		os.Unsetenv("CONFIG_AWS_ACCOUNT_ID")
		os.Unsetenv("CONFIG_BUCKET_NAME_SUFFIX")
	})

	t.Run("load Database from env", func(t *testing.T) {
		os.Setenv("CONFIG_NETWORK_NAME", "test_network")
		os.Setenv("CONFIG_GSHEET_ID", "test_gsheet_id")
		os.Setenv("DELEGATION_WHITELIST_LIST", "test_list")
		os.Setenv("DELEGATION_WHITELIST_COLUMN", "test_column")
		os.Setenv("CONFIG_DATABASE_CONNECTION_STRING", "test_connection_string")
		os.Setenv("CONFIG_DATABASE_TYPE", "test_type")

		config := LoadEnv(mockLogger)
		if config.Database == nil || config.Database.ConnectionString != "test_connection_string" {
			t.Error("Failed to load DB configs from environment variables")
		}

		// Cleanup
		os.Unsetenv("CONFIG_NETWORK_NAME")
		os.Unsetenv("CONFIG_GSHEET_ID")
		os.Unsetenv("DELEGATION_WHITELIST_LIST")
		os.Unsetenv("DELEGATION_WHITELIST_COLUMN")
		os.Unsetenv("CONFIG_DATABASE_CONNECTION_STRING")
		os.Unsetenv("CONFIG_DATABASE_TYPE")
	})

	t.Run("load Filesystem from env", func(t *testing.T) {
		os.Setenv("CONFIG_NETWORK_NAME", "test_network")
		os.Setenv("CONFIG_GSHEET_ID", "test_gsheet_id")
		os.Setenv("DELEGATION_WHITELIST_LIST", "test_list")
		os.Setenv("DELEGATION_WHITELIST_COLUMN", "test_column")
		os.Setenv("CONFIG_FILESYSTEM_PATH", "test_path")

		config := LoadEnv(mockLogger)
		if config.LocalFileSystem == nil || config.LocalFileSystem.Path != "test_path" {
			t.Error("Failed to load Filesystem configs from environment variables")
		}

		// Cleanup
		os.Unsetenv("CONFIG_NETWORK_NAME")
		os.Unsetenv("CONFIG_GSHEET_ID")
		os.Unsetenv("DELEGATION_WHITELIST_LIST")
		os.Unsetenv("DELEGATION_WHITELIST_COLUMN")
		os.Unsetenv("CONFIG_FILESYSTEM_PATH")
	})

	t.Run("multiple configs error from env", func(t *testing.T) {
		// Set env variables for both AWS and Database
		os.Setenv("AWS_ACCESS_KEY_ID", "test_access_key_id")
		os.Setenv("CONFIG_DATABASE_CONNECTION_STRING", "test_connection_string")

		LoadEnv(mockLogger)
		if mockLogger.lastMessage != "Error: You can only provide one of Aws, Database, or LocalFileSystem configurations." {
			t.Error("Expected to get an error for multiple configs but didn't.")
		}

		// Cleanup
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("CONFIG_DATABASE_CONNECTION_STRING")
	})

	t.Run("multiple configs error from file", func(t *testing.T) {
		// Create a temporary config file
		fileContent := `
			{
				"network_name": "test_network",
				"gsheet_id": "test_gsheet_id",
				"delegation_whitelist_list": "test_list",
				"delegation_whitelist_column": "test_column",
				"aws": {
					"account_id": "test_account_id",
					"bucket_name_suffix": "test_suffix",
					"region": "test_region",
					"access_key_id": "test_access_key_id",
					"secret_access_key": "test_secret_access_key"
				},
				"database": {
					"connection_string": "test_connection_string",
					"type": "test_type"
				}
			}
			`
		tmpFile := "/tmp/test_config.json"
		os.WriteFile(tmpFile, []byte(fileContent), 0644)
		os.Setenv("CONFIG_FILE", tmpFile)
		LoadEnv(mockLogger)
		if mockLogger.lastMessage != "Error: You can only provide one of Aws, Database, or LocalFileSystem configurations." {
			t.Error("Expected to get an error for multiple configs but didn't.")
		}
	})
}