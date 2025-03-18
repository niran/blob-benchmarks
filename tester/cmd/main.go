package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/log"
	"github.com/niran/blob-benchmarks/tester"
	"github.com/urfave/cli/v3"
)

func main() {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, true)))

	cmd := &cli.Command{
		Name:   "blob-benchmarks",
		Usage:  "Determine the networking limits of a reproducible Ethereum network simulation",
		Action: minBandwidth,
		/*
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "enclave",
					Aliases:  []string{"e"},
					Usage:    "The name of a running enclave to use",
					Required: false,
				},
			},
		*/
		Commands: []*cli.Command{
			{
				Name:  "min-bandwidth",
				Usage: "Determine the minimum bandwidth required for a given number of blobs per block",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "blobs",
						Aliases: []string{"b"},
						Usage:   "The number of blobs per block",
						Value:   6,
					},
					&cli.IntFlag{
						Name:    "bandwidth",
						Aliases: []string{"bw"},
						Usage:   "The initial bandwidth in megabits per second",
						Value:   50,
					},
					&cli.IntFlag{
						Name:    "delta",
						Aliases: []string{"d"},
						Usage:   "The percentage to decrease the bandwidth by each iteration",
						Value:   50,
					},
				},
				Action: minBandwidth,
			},
			{
				Name:   "max-blobs",
				Usage:  "Determine the maximum number of blobs per block that can be sustained by a node given a target bandwidth",
				Action: maxBlobs,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "bandwidth",
						Aliases: []string{"bw"},
						Usage:   "The target node's bandwidth in megabits per second",
						Value:   50,
					},
					&cli.IntFlag{
						Name:    "blobs",
						Aliases: []string{"b"},
						Usage:   "The initial number of blobs per block",
						Value:   6,
					},
					&cli.IntFlag{
						Name:    "delta",
						Aliases: []string{"d"},
						Usage:   "The percentage to increase the blob count by each iteration",
						Value:   100,
					},
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Crit("Application failed", "error", err)
	}
}

func minBandwidth(ctx context.Context, cmd *cli.Command) error {
	log.Info("Starting blob-benchmarks")

	enclaveContext, err := tester.GetOnlyEnclaveContext(ctx)
	if err != nil {
		return err
	}

	log.Info("Retrieved enclave context", "name", enclaveContext.GetEnclaveName())

	// TODO: If we created an enclave, defer its deletion.
	createdEnclave := false
	if createdEnclave {
		defer tester.CleanupEnclave(enclaveContext)
	}

	testDoneChannel := make(chan struct{})
	go func() {
		test := tester.NewMinBandwidthTest(enclaveContext, uint(cmd.Int("blobs")), uint(cmd.Int("bandwidth")), uint(cmd.Int("delta")))
		err = test.Run(testDoneChannel)
		if err != nil {
			log.Crit("Test failed", "error", err)
		}
	}()

	defer func() {
		log.Info("Cleaning up bandwidth controls...")
		service, err := tester.GetServiceUnderTest(enclaveContext)
		if err != nil {
			log.Error("Failed to get service under test", "error", err)
		}

		tester.RemoveBandwidthControls(service)
	}()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-interruptChannel:
		log.Info("Stopping test...")
	case <-testDoneChannel:
		log.Info("Test completed.")
	}

	return nil
}

func maxBlobs(ctx context.Context, cmd *cli.Command) error {
	return nil
}
