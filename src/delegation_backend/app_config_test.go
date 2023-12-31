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

func TestGetEnvChecked(t *testing.T) {
	// Reset environment variables
	os.Clearenv()

	// Mock logger instance
	mockLogger := &MockLogger{}

	// Case 1: Environment variable is set
	os.Setenv("TEST_VARIABLE", "testvalue")
	value := getEnvChecked("TEST_VARIABLE", mockLogger)
	if value != "testvalue" {
		t.Errorf("Expected 'testvalue', got '%s'", value)
	}

	// Case 2: Environment variable is NOT set
	getEnvChecked("MISSING_VARIABLE", mockLogger)
	if !(mockLogger.lastMessage == "missing MISSING_VARIABLE environment variable") {
		t.Error("Expected Fatalf to be called due to missing environment variable")
	}

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

	t.Run("load AwsKeyspaces from config file", func(t *testing.T) {
		// Create a temporary config file
		fileContent := `
			{
				"network_name": "test_network",
				"gsheet_id": "test_gsheet_id",
				"delegation_whitelist_list": "test_list",
				"delegation_whitelist_column": "test_column",
				"aws_keyspaces": {
					"keyspace": "test_keyspace",
					"region": "test_region",
					"access_key_id": "test_access_key_id",
					"secret_access_key": "test_secret_access_key",
					"ssl_certificate_path": "test_ssl_certificate_path"
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
		if config.AwsKeyspaces == nil {
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
		os.Setenv("AWS_REGION", "test_region")
		os.Setenv("AWS_ACCOUNT_ID", "test_account_id")
		os.Setenv("AWS_BUCKET_NAME_SUFFIX", "test_suffix")

		config := LoadEnv(mockLogger)
		if config.Aws == nil || config.Aws.AccessKeyId != "test_access_key_id" {
			t.Error("Failed to load AWS configs from environment variables")
		}

		// Cleanup
		os.Clearenv()
	})

	t.Run("load AwsKeyspaces from env", func(t *testing.T) {
		os.Setenv("CONFIG_NETWORK_NAME", "test_network")
		os.Setenv("CONFIG_GSHEET_ID", "test_gsheet_id")
		os.Setenv("DELEGATION_WHITELIST_LIST", "test_list")
		os.Setenv("DELEGATION_WHITELIST_COLUMN", "test_column")
		os.Setenv("AWS_ACCESS_KEY_ID", "test_access_key_id")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test_secret_access_key")
		os.Setenv("AWS_REGION", "test_region")
		os.Setenv("AWS_KEYSPACE", "test_keyspace")
		os.Setenv("AWS_SSL_CERTIFICATE_PATH", "test_ssl_certificate_path")

		config := LoadEnv(mockLogger)
		if config.AwsKeyspaces == nil || config.AwsKeyspaces.Keyspace != "test_keyspace" {
			t.Error("Failed to load DB configs from environment variables")
		}

		// Cleanup
		os.Clearenv()
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
		if config.DelegationWhitelistDisabled != false {
			t.Error("Expected DelegationWhitelistDisabled to be false but got true")
		}

		// Cleanup
		os.Clearenv()
	})

	t.Run("multiple configs from env", func(t *testing.T) {
		os.Clearenv()
		// Set env variables for both AWS and Database
		os.Setenv("CONFIG_NETWORK_NAME", "test_network")
		os.Setenv("CONFIG_GSHEET_ID", "test_gsheet_id")
		os.Setenv("DELEGATION_WHITELIST_LIST", "test_list")
		os.Setenv("DELEGATION_WHITELIST_COLUMN", "test_column")

		os.Setenv("AWS_ACCESS_KEY_ID", "test_access_key_id")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test_secret_access_key")
		os.Setenv("AWS_REGION", "test_region")
		os.Setenv("AWS_ACCOUNT_ID", "test_account_id")
		os.Setenv("AWS_BUCKET_NAME_SUFFIX", "test_suffix")

		os.Setenv("AWS_KEYSPACE", "test_keyspace")
		os.Setenv("AWS_SSL_CERTIFICATE_PATH", "test_ssl_cert")

		config := LoadEnv(mockLogger)
		if config.AwsKeyspaces == nil || config.AwsKeyspaces.Keyspace != "test_keyspace" {
			t.Error("Failed to load DB configs from environment variables")
		}
		if config.Aws == nil || config.Aws.AccessKeyId != "test_access_key_id" {
			t.Error("Failed to load AWS configs from environment variables")
		}

		// Cleanup
		os.Clearenv()
	})

	t.Run("multiple configs from file", func(t *testing.T) {
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
				"filesystem": {
					"path": "test_path"
				},
				"aws_keyspaces": {
					"keyspace": "test_keyspace",
					"region": "test_region",
					"access_key_id": "test_access_key_id",
					"secret_access_key": "test_secret_access_key",
					"ssl_certificate_path": "test_ssl_certificate_path"
				}
			}
			`
		tmpFile := "/tmp/test_config.json"
		os.WriteFile(tmpFile, []byte(fileContent), 0644)
		os.Setenv("CONFIG_FILE", tmpFile)
		config := LoadEnv(mockLogger)
		if config.AwsKeyspaces == nil || config.AwsKeyspaces.Keyspace != "test_keyspace" {
			t.Error("Failed to load DB configs from file")
		}
		if config.Aws == nil || config.Aws.AccessKeyId != "test_access_key_id" {
			t.Error("Failed to load AWS configs from file")
		}
		if config.LocalFileSystem == nil || config.LocalFileSystem.Path != "test_path" {
			t.Error("Failed to load Filesystem configs from file")
		}

	})

	t.Run("delegation whitelist disabled - env", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CONFIG_NETWORK_NAME", "test_network")
		os.Setenv("DELEGATION_WHITELIST_DISABLED", "1")
		os.Setenv("CONFIG_FILESYSTEM_PATH", "test_path")

		config := LoadEnv(mockLogger)
		if config.DelegationWhitelistDisabled != true {
			t.Error("Expected DelegationWhitelistDisabled to be true but got false")
		}

		// Cleanup
		os.Clearenv()
	})

	t.Run("delegation whitelist disabled - file", func(t *testing.T) {
		os.Clearenv()
		// Create a temporary config file
		fileContent := `
			{
				"network_name": "test_network",
				"delegation_whitelist_disabled": true,
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
		if config.DelegationWhitelistDisabled != true {
			t.Error("Expected DelegationWhitelistDisabled to be true but got false")
		}
	})
}
