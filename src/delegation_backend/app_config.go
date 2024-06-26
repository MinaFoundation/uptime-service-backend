package delegation_backend

import (
	"encoding/json"
	"os"
	"strconv"

	logging "github.com/ipfs/go-log/v2"
)

func GetAWSBucketName(config AppConfig) string {
	if config.Aws != nil {
		return config.Aws.AccountId + "-" + config.Aws.BucketNameSuffix
	}
	return "" // return empty in case AWSConfig is nil
}

func LoadEnv(log logging.EventLogger) AppConfig {
	var config AppConfig

	configFile := os.Getenv("CONFIG_FILE")
	if configFile != "" {
		file, err := os.Open(configFile)
		if err != nil {
			log.Fatalf("Error loading config file: %s", err)
		}
		defer file.Close()
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		if err != nil {
			log.Fatalf("Error decoding config file: %s", err)
		}
		// Set AWS credentials from config file in case we are using AWS S3 or AWS Keyspaces
		if config.Aws != nil {
			os.Setenv("AWS_ACCESS_KEY_ID", config.Aws.AccessKeyId)
			os.Setenv("AWS_SECRET_ACCESS_KEY", config.Aws.SecretAccessKey)
		}
	} else {
		// networkName is used as part of the S3 bucket path and influences networkId
		// networkName = "mainnet" will result in networkId = 1 else networkId = 0 and this influeces verifySignature
		networkName := getEnvChecked("CONFIG_NETWORK_NAME", log)
		verifySignatureDisabled := boolEnvChecked("VERIFY_SIGNATURE_DISABLED", log)

		delegationWhitelistDisabled := boolEnvChecked("DELEGATION_WHITELIST_DISABLED", log)
		var gsheetId, delegationWhitelistList, delegationWhitelistColumn string
		if delegationWhitelistDisabled {
			// If delegation whitelist is disabled, we don't need to load related environment variables
			// just loading them from env in case they are set, but they won't be used
			gsheetId = os.Getenv("CONFIG_GSHEET_ID")
			delegationWhitelistList = os.Getenv("DELEGATION_WHITELIST_LIST")
			delegationWhitelistColumn = os.Getenv("DELEGATION_WHITELIST_COLUMN")
		} else {
			// If delegation whitelist is enabled, we need to load related environment variables
			// program will terminate if any of them is missing
			gsheetId = getEnvChecked("CONFIG_GSHEET_ID", log)
			delegationWhitelistList = getEnvChecked("DELEGATION_WHITELIST_LIST", log)
			delegationWhitelistColumn = getEnvChecked("DELEGATION_WHITELIST_COLUMN", log)
		}

		// AWS configurations
		if bucketNameSuffix := os.Getenv("AWS_BUCKET_NAME_SUFFIX"); bucketNameSuffix != "" {
			// accessKeyId, secretAccessKey are not mandatory for production set up
			accessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
			secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
			awsRegion := getEnvChecked("AWS_REGION", log)
			awsAccountId := getEnvChecked("AWS_ACCOUNT_ID", log)
			bucketNameSuffix := getEnvChecked("AWS_BUCKET_NAME_SUFFIX", log)

			config.Aws = &AwsConfig{
				AccountId:        awsAccountId,
				BucketNameSuffix: bucketNameSuffix,
				Region:           awsRegion,
				AccessKeyId:      accessKeyId,
				SecretAccessKey:  secretAccessKey,
			}
		}

		// AWSKeyspace/Cassandra configurations
		if keyspace := os.Getenv("AWS_KEYSPACE"); keyspace != "" {
			awsKeyspace := getEnvChecked("AWS_KEYSPACE", log)
			sslCertificatePath := getEnvChecked("AWS_SSL_CERTIFICATE_PATH", log)

			//service level connection
			cassandraHost := os.Getenv("CASSANDRA_HOST")
			cassandraPortStr := os.Getenv("CASSANDRA_PORT")
			cassandraPort, err := strconv.Atoi(cassandraPortStr)
			if err != nil {
				cassandraPort = 9142
			}
			cassandraUsername := os.Getenv("CASSANDRA_USERNAME")
			cassandraPassword := os.Getenv("CASSANDRA_PASSWORD")

			//aws keyspaces connection
			awsRegion := os.Getenv("AWS_REGION")

			// if webIdentityTokenFile, roleSessionName and roleArn are set,
			// we are using AWS STS to assume a role and get temporary credentials
			// if they are not set, we are using AWS IAM user credentials
			webIdentityTokenFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
			roleSessionName := os.Getenv("AWS_ROLE_SESSION_NAME")
			roleArn := os.Getenv("AWS_ROLE_ARN")
			// accessKeyId, secretAccessKey are not mandatory for production set up
			accessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
			secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

			config.AwsKeyspaces = &AwsKeyspacesConfig{
				Keyspace:             awsKeyspace,
				CassandraHost:        cassandraHost,
				CassandraPort:        cassandraPort,
				CassandraUsername:    cassandraUsername,
				CassandraPassword:    cassandraPassword,
				Region:               awsRegion,
				AccessKeyId:          accessKeyId,
				SecretAccessKey:      secretAccessKey,
				WebIdentityTokenFile: webIdentityTokenFile,
				RoleSessionName:      roleSessionName,
				RoleArn:              roleArn,
				SSLCertificatePath:   sslCertificatePath,
			}
		}

		// LocalFileSystem configurations
		if path := os.Getenv("CONFIG_FILESYSTEM_PATH"); path != "" {
			config.LocalFileSystem = &LocalFileSystemConfig{
				Path: path,
			}
		}

		// PostgreSQL configurations
		if postgresHost := os.Getenv("POSTGRES_HOST"); postgresHost != "" {
			postgresUser := getEnvChecked("POSTGRES_USER", log)
			postgresPassword := getEnvChecked("POSTGRES_PASSWORD", log)
			postgresDBName := getEnvChecked("POSTGRES_DB", log)
			postgresPort, err := strconv.Atoi(getEnvChecked("POSTGRES_PORT", log))
			if err != nil {
				log.Fatalf("Error parsing POSTGRES_PORT: %v", err)
			}
			postgresSSLMode := os.Getenv("POSTGRES_SSLMODE")
			if postgresSSLMode == "" {
				postgresSSLMode = "require"
			}

			config.PostgreSQL = &PostgreSQLConfig{
				Host:     postgresHost,
				Port:     postgresPort,
				User:     postgresUser,
				Password: postgresPassword,
				DBName:   postgresDBName,
				SSLMode:  postgresSSLMode,
			}
		}

		config.NetworkName = networkName
		config.GsheetId = gsheetId
		config.DelegationWhitelistList = delegationWhitelistList
		config.DelegationWhitelistColumn = delegationWhitelistColumn
		config.DelegationWhitelistDisabled = delegationWhitelistDisabled
		config.VerifySignatureDisabled = verifySignatureDisabled
	}

	return config
}

func getEnvChecked(variable string, log logging.EventLogger) string {
	value := os.Getenv(variable)
	if value == "" {
		log.Fatalf("missing %s environment variable", variable)
	}
	return value
}

func boolEnvChecked(variable string, log logging.EventLogger) bool {
	value := os.Getenv(variable)
	switch value {
	case "1":
		return true
	case "0":
		return false
	case "":
		return false
	default:
		log.Fatalf("%s, if set, should be either 0 or 1!", variable)
		return false
	}
}

type AwsConfig struct {
	AccountId        string `json:"account_id"`
	BucketNameSuffix string `json:"bucket_name_suffix"`
	Region           string `json:"region"`
	AccessKeyId      string `json:"access_key_id"`
	SecretAccessKey  string `json:"secret_access_key"`
}

type AwsKeyspacesConfig struct {
	Keyspace             string `json:"keyspace"`
	CassandraHost        string `json:"cassandra_host"`
	CassandraPort        int    `json:"cassandra_port"`
	CassandraUsername    string `json:"cassandra_username,omitempty"`
	CassandraPassword    string `json:"cassandra_password,omitempty"`
	Region               string `json:"region,omitempty"`
	AccessKeyId          string `json:"access_key_id,omitempty"`
	SecretAccessKey      string `json:"secret_access_key,omitempty"`
	WebIdentityTokenFile string `json:"web_identity_token_file,omitempty"`
	RoleSessionName      string `json:"role_session_name,omitempty"`
	RoleArn              string `json:"role_arn,omitempty"`
	SSLCertificatePath   string `json:"ssl_certificate_path"`
}

type LocalFileSystemConfig struct {
	Path string `json:"path"`
}

type PostgreSQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

type AppConfig struct {
	NetworkName                 string                 `json:"network_name"`
	GsheetId                    string                 `json:"gsheet_id"`
	DelegationWhitelistList     string                 `json:"delegation_whitelist_list"`
	DelegationWhitelistColumn   string                 `json:"delegation_whitelist_column"`
	DelegationWhitelistDisabled bool                   `json:"delegation_whitelist_disabled,omitempty"`
	VerifySignatureDisabled     bool                   `json:"verify_signature_disabled,omitempty"`
	Aws                         *AwsConfig             `json:"aws,omitempty"`
	AwsKeyspaces                *AwsKeyspacesConfig    `json:"aws_keyspaces,omitempty"`
	LocalFileSystem             *LocalFileSystemConfig `json:"filesystem,omitempty"`
	PostgreSQL                  *PostgreSQLConfig      `json:"postgresql,omitempty"`
}
