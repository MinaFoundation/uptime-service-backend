package integration_tests

import (
	"log"
	"testing"
)

func init() {
	log.Println("Integration tests setup")
	// err := encodeSecrets()
	// if err != nil {
	// 	log.Fatalf("Failed to encode secrets: %v", err)
	// }
	err := decodeSecrets()
	if err != nil {
		log.Fatalf("Failed to decode secrets: %v", err)
	}
}

func TestIntegrationBP_Uptime_S3(t *testing.T) {
	creds := getAWSCreds()
	config := getAppConfig()

	defer miniminaNetworkDelete(config.NetworkName)
	defer emptyIntegrationTestFolder(creds, config)

	err := emptyIntegrationTestFolder(creds, config)
	if err != nil {
		t.Fatalf("Failed to empty the integration_test folder: %v", err)
	}

	miniminaNetworkCreate(config.NetworkName)
	miniminaNetworkStart(config.NetworkName)

	err = waitUntilS3BucketHasBlocksAndSubmissions(creds, config)
	if err != nil {
		t.Fatalf("Failed to wait until S3 bucket is not empty: %v", err)
	}

}
