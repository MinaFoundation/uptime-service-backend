package main

import (
    dg "block_producers_uptime/delegation_backend"
    itn "block_producers_uptime/itn_uptime_analyzer"
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go/aws"
    logging "github.com/ipfs/go-log/v2"
)

func main() {
    // Set up sync period of type int representing minutes
    syncPeriod := 15

    // Setting up logging for application
    logging.SetupLogging(logging.Config{
        Format: logging.JSONOutput,
        Stderr: true,
        Stdout: false,
        Level:  logging.LevelDebug,
        File:   "",
    })
    log := logging.Logger("itn availability script")

    // Empty context object and initializing memory for application
    ctx := context.Background()

    // Load environment variables
    appCfg := itn.LoadEnv(log)

    awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(appCfg.Aws.Region))
    if err != nil {
        log.Fatalf("Error loading AWS configuration: %v\n", err)
    }

    app := new(dg.App)
    app.Log = log
    client := s3.NewFromConfig(awsCfg)

    awsctx := dg.AwsContext{Client: client, BucketName: aws.String(itn.GetBucketName(appCfg)), Prefix: appCfg.NetworkName, Context: ctx, Log: log}

    if appCfg.IgnoreIPs {
        fmt.Printf("Period start; %v\nPeriod end; %v\n", appCfg.Period.Start, appCfg.Period.End)
        fmt.Printf("Interval; %v\npublic key; uptime (%%)\n", appCfg.Period.Interval)
    } else {
        fmt.Printf("Period start; %v;\nPeriod end; %v;\n", appCfg.Period.Start, appCfg.Period.End)
        fmt.Printf("Interval; %v;\npublic key; public ip; uptime (%%)\n", appCfg.Period.Interval)
    }

    identities := itn.CreateIdentities(appCfg, awsctx, log)
    // Go over identities and calculate uptime
    for _, identity := range identities {
        identity.GetUptime(appCfg, awsctx, log, syncPeriod)
        if appCfg.IgnoreIPs {
            fmt.Printf("%s; %s\n", identity.PublicKey, *identity.Uptime)
        } else {
            fmt.Printf("%s; %s; %s\n", identity.PublicKey, identity.PublicIp, *identity.Uptime)
        }
    }
}
