package delegation_backend

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sigv4-auth-cassandra-gocql-driver-plugin/sigv4"
	"github.com/gocql/gocql"
	"github.com/golang-migrate/migrate/v4"
	cassandra "github.com/golang-migrate/migrate/v4/database/cassandra"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	logging "github.com/ipfs/go-log/v2"
)

// InitializeKeyspaceSession creates a new gocql session for Amazon Keyspaces using the provided configuration.
func InitializeKeyspaceSession(config *AwsKeyspacesConfig) (*gocql.Session, error) {
	var cluster *gocql.ClusterConfig

	var endpoint string
	if config.CassandraHost == "" {
		if config.Region == "" {
			return nil, fmt.Errorf("AWS_REGION is required when CASSANDRA_HOST is not set")
		}
		endpoint = "cassandra." + config.Region + ".amazonaws.com"
	} else {
		endpoint = config.CassandraHost
	}

	cluster = gocql.NewCluster(endpoint)
	cluster.Keyspace = config.Keyspace

	var port int
	if config.CassandraPort != 0 {
		port = config.CassandraPort
	} else {
		port = 9142
	}
	cluster.Port = port

	if config.CassandraUsername != "" && config.CassandraPassword != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: config.CassandraUsername,
			Password: config.CassandraPassword}
	} else {
		var err error
		cluster.Authenticator, err = sigv4Authentication(config)
		if err != nil {
			return nil, fmt.Errorf("could not create SigV4 authenticator: %w", err)
		}
	}

	cluster.SslOpts = &gocql.SslOptions{
		CaPath: config.SSLCertificatePath,

		EnableHostVerification: false,
	}

	cluster.Consistency = gocql.LocalQuorum
	cluster.DisableInitialHostLookup = false
	cluster.RetryPolicy = &gocql.ExponentialBackoffRetryPolicy{NumRetries: 10, Min: 100 * time.Millisecond, Max: 10 * time.Second}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("could not create Cassandra session: %w", err)
	}

	return session, nil
}

func sigv4Authentication(config *AwsKeyspacesConfig) (sigv4.AwsAuthenticator, error) {
	auth := sigv4.NewAwsAuthenticator()
	if config.RoleSessionName != "" && config.RoleArn != "" && config.WebIdentityTokenFile != "" {
		// If role-related env variables are set, use temporary credentials
		tokenBytes, err := os.ReadFile(config.WebIdentityTokenFile)
		if err != nil {
			return auth, fmt.Errorf("error reading web identity token file: %w", err)
		}
		webIdentityToken := string(tokenBytes)

		awsSession, err := session.NewSession(&aws.Config{Region: aws.String(config.Region)})
		if err != nil {
			return auth, fmt.Errorf("error creating AWS session: %w", err)
		}

		stsSvc := sts.New(awsSession)
		creds, err := stsSvc.AssumeRoleWithWebIdentity(&sts.AssumeRoleWithWebIdentityInput{
			RoleArn:          &config.RoleArn,
			RoleSessionName:  &config.RoleSessionName,
			WebIdentityToken: &webIdentityToken,
		})
		if err != nil {
			return auth, fmt.Errorf("unable to assume role: %w", err)
		}

		auth.AccessKeyId = *creds.Credentials.AccessKeyId
		auth.SecretAccessKey = *creds.Credentials.SecretAccessKey
		auth.SessionToken = *creds.Credentials.SessionToken
		auth.Region = config.Region
	} else {
		// Otherwise, use credentials from the config
		auth.AccessKeyId = config.AccessKeyId
		auth.SecretAccessKey = config.SecretAccessKey
		auth.Region = config.Region
	}
	return auth, nil
}

type KeyspaceContext struct {
	Session  *gocql.Session
	Keyspace string
	Context  context.Context
	Log      *logging.ZapEventLogger
}

// calculateShard returns the shard number for a given submission time.
// 0-599 are the possible shard numbers, each representing a 144-second interval within 24h.
// shard = (3600 * hour + 60 * minute + second) // 144
func calculateShard(submittedAt time.Time) int {
	hour := submittedAt.Hour()
	minute := submittedAt.Minute()
	second := submittedAt.Second()
	return (3600*hour + 60*minute + second) / 144
}

// Estimate the size of the raw block in bytes.
// In Go, len() returns the number of bytes in a slice, which should suffice for a rough estimation.
func calculateBlockSize(rawBlock []byte) int {
	return len(rawBlock)
}

// Insert a submission into the Keyspaces database
func (kc *KeyspaceContext) insertSubmission(submission *Submission) error {
	return ExponentialBackoff(func() error {
		if submission.RawBlock == nil {
			kc.Log.Error("KeyspaceSave: Block is missing in the submission, which is not expected, but inserting without raw_block")
			if err := kc.insertSubmissionWithoutRawBlock(submission); err != nil {
				return err
			}
		} else if calculateBlockSize(submission.RawBlock) > MAX_BLOCK_SIZE {
			kc.Log.Infof("KeyspaceSave: Block too large (%d bytes), inserting without raw_block", calculateBlockSize(submission.RawBlock))
			if err := kc.insertSubmissionWithoutRawBlock(submission); err != nil {
				return err
			}
		} else {
			if err := kc.insertSubmissionWithRawBlock(submission); err != nil {
				return err
			}

		}

		return nil
	}, maxRetries, initialBackoff)
}

func (kc *KeyspaceContext) insertSubmissionWithoutRawBlock(submission *Submission) error {
	query := "INSERT INTO " + kc.Keyspace + ".submissions (submitted_at_date, shard, submitted_at, submitter, remote_addr, peer_id, snark_work, block_hash, created_at, graphql_control_port, built_with_commit_sha) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	values := []interface{}{
		submission.SubmittedAtDate,
		calculateShard(submission.SubmittedAt),
		submission.SubmittedAt,
		submission.Submitter,
		submission.RemoteAddr,
		submission.PeerId,
		submission.SnarkWork,
		submission.BlockHash,
		submission.CreatedAt,
		submission.GraphqlControlPort,
		submission.BuiltWithCommitSha,
	}
	return kc.Session.Query(query, values...).Exec()
}

func (kc *KeyspaceContext) insertSubmissionWithRawBlock(submission *Submission) error {
	query := "INSERT INTO " + kc.Keyspace + ".submissions (submitted_at_date, shard, submitted_at, submitter, remote_addr, peer_id, snark_work, block_hash, created_at, graphql_control_port, built_with_commit_sha, raw_block) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	values := []interface{}{
		submission.SubmittedAtDate,
		calculateShard(submission.SubmittedAt),
		submission.SubmittedAt,
		submission.Submitter,
		submission.RemoteAddr,
		submission.PeerId,
		submission.SnarkWork,
		submission.BlockHash,
		submission.CreatedAt,
		submission.GraphqlControlPort,
		submission.BuiltWithCommitSha,
		submission.RawBlock,
	}
	return kc.Session.Query(query, values...).Exec()
}

// KeyspaceSave saves the provided objects into Amazon Keyspaces.
func (kc *KeyspaceContext) KeyspaceSave(objs ObjectsToSave) {
	submissionToSave, err := objectToSaveToSubmission(objs, kc.Log)
	if err != nil {
		kc.Log.Errorf("KeyspaceSave: Error preparing submission for saving: %v", err)
		return
	}
	kc.Log.Infof("KeyspaceSave: Saving submission for block: %v, submitter: %v, submitted_at: %v", submissionToSave.BlockHash, submissionToSave.Submitter, submissionToSave.SubmittedAt)
	if err := kc.insertSubmission(submissionToSave); err != nil {
		kc.Log.Errorf("KeyspaceSave: Error saving submission to Keyspaces: %v", err)
	}
}

func createSchemaMigrationsTableIfNotExists(session *gocql.Session, keyspace string) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.schema_migrations (version bigint PRIMARY KEY, dirty boolean);`, keyspace)
	operation := func() error {
		return session.Query(query).Exec()
	}

	return ExponentialBackoff(operation, 5, 1*time.Second)
}

func DropAllTables(config *AwsKeyspacesConfig) error {
	log.Print("Dropping all tables...")
	operation := func() error {
		session, err := InitializeKeyspaceSession(config)
		if err != nil {
			return fmt.Errorf("could not initialize Cassandra session: %w", err)
		}
		query := fmt.Sprintf(`SELECT table_name FROM system_schema.tables WHERE keyspace_name = '%s';`, config.Keyspace)
		iter := session.Query(query).Iter()
		var tableName string
		for iter.Scan(&tableName) {
			query = fmt.Sprintf(`DROP TABLE %s.%s;`, config.Keyspace, tableName)
			err := session.Query(query).Exec()
			if err != nil {
				return fmt.Errorf("could not drop table %s: %w", tableName, err)
			}
		}
		if err := iter.Close(); err != nil {
			return fmt.Errorf("could not close iterator: %w", err)
		}

		return nil
	}

	return ExponentialBackoff(operation, maxRetries, initialBackoff)

}

// MigrationUp applies all up migrations.
func MigrationUp(config *AwsKeyspacesConfig, migrationPath string) error {
	log.Print("Running database migration Up...")
	session, err := InitializeKeyspaceSession(config)
	if err != nil {
		return fmt.Errorf("could not initialize Cassandra session: %w", err)
	}
	defer session.Close()

	// Check if the schema_migrations table exists, create if not
	err = createSchemaMigrationsTableIfNotExists(session, config.Keyspace)
	if err != nil {
		return fmt.Errorf("could not create schema_migrations table: %w", err)
	}

	//run migrations
	operation := func() error {
		driver, err := cassandra.WithInstance(session, &cassandra.Config{
			KeyspaceName: config.Keyspace,
		})
		if err != nil {
			return fmt.Errorf("could not create Cassandra migration driver: %w", err)
		}

		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", migrationPath),
			config.Keyspace, driver)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("an error occurred while applying migrations: %w", err)
		}
		return nil
	}

	return ExponentialBackoff(operation, 10, 1*time.Second)

}

// MigrationDown rolls back all migrations.
func MigrationDown(config *AwsKeyspacesConfig, migrationPath string) error {
	log.Print("Running database migration Down...")
	session, err := InitializeKeyspaceSession(config)
	if err != nil {
		return err
	}
	defer session.Close()

	operation := func() error {
		driver, err := cassandra.WithInstance(session, &cassandra.Config{
			KeyspaceName: config.Keyspace,
		})
		if err != nil {
			return fmt.Errorf("could not create Cassandra migration driver: %w", err)
		}

		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", migrationPath),
			config.Keyspace, driver)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("an error occurred while rolling back migrations: %w", err)
		}

		return nil
	}

	return ExponentialBackoff(operation, maxRetries, initialBackoff)
}
