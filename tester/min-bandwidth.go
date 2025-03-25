package tester

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/niran/blob-benchmarks/tester/checks"
	"github.com/pkg/errors"
)

const (
	slotsPerEpoch = 32
	slotDuration  = 12 * time.Second
)

type MinBandwidthTestConfig struct {
	enclaveContext *enclaves.EnclaveContext
	blobsPerBlock  uint
	bandwidth      uint
	minBandwidth   uint
	delta          uint
}

type MinBandwidthTest struct {
	cfg              MinBandwidthTestConfig
	currentBandwidth uint
	startTime        time.Time
}

func NewMinBandwidthTest(enclaveContext *enclaves.EnclaveContext, blobsPerBlock uint, bandwidth uint, minBandwidth uint, delta uint) *MinBandwidthTest {
	return &MinBandwidthTest{
		cfg: MinBandwidthTestConfig{
			enclaveContext: enclaveContext,
			blobsPerBlock:  blobsPerBlock,
			bandwidth:      bandwidth,
			minBandwidth:   minBandwidth,
			delta:          delta,
		},
		currentBandwidth: bandwidth,
		startTime:        time.Now(),
	}
}

func NewMinBandwidthTestForOnlyEnclave(ctx context.Context, blobsPerBlock uint, bandwidth uint, minBandwidth uint, delta uint) (*MinBandwidthTest, error) {
	enclaveContext, err := GetOnlyEnclaveContext(ctx)
	if err != nil {
		return nil, err
	}

	return NewMinBandwidthTest(enclaveContext, blobsPerBlock, bandwidth, minBandwidth, delta), nil
}

func (t *MinBandwidthTest) getElapsedEpochs() uint {
	timeSinceStart := time.Since(t.startTime)
	slotsSinceStart := uint(timeSinceStart / slotDuration)
	return slotsSinceStart / slotsPerEpoch
}

func (t *MinBandwidthTest) Run(doneChannel chan struct{}) error {
	grafanaBaseURL, grafanaToken, datasourceID, err := GetGrafanaConfig(t.cfg.enclaveContext)
	if err != nil {
		return errors.Wrap(err, "failed to get grafana config")
	}

	runner, err := checks.SetupRunner(grafanaBaseURL, grafanaToken, datasourceID)
	if err != nil {
		return errors.Wrap(err, "failed to setup runner")
	}

	if err := runner.RunChecks(context.Background()); err != nil {
		log.Error("Failed to run checks", "error", err)
	} else {
		log.Info("Check results", "results", runner.GetResults())
		log.Info("Check analysis", "analysis", runner.GetAnalysis())
		// log.Info("Check logs", "logs", runner.GetLog().GetBuffer().String())
	}

	// Get the service for the node whose bandwidth we want to limit.
	service, err := GetServiceUnderTest(t.cfg.enclaveContext)
	if err != nil {
		return errors.Wrap(err, "failed to get service under test")
	}

	// Install the tc command.
	if err := InstallTcCommand(service); err != nil {
		return errors.Wrap(err, "failed to install tc command")
	}

	// Remove any existing bandwidth controls.
	if err := RemoveBandwidthControls(service); err != nil {
		log.Info("No existing bandwidth controls seem to be set, continuing...", "message", err)
	}

	// Set the download bandwidth to a reasonable fixed value.
	if err := SetDownloadBandwidthControl(service, 100_000_000); err != nil {
		return errors.Wrap(err, "failed to set download bandwidth control")
	}

	// Set the upload bandwith to a starting point for the tests.
	if err := SetUploadBandwidthControl(service, t.currentBandwidth); err != nil {
		return errors.Wrap(err, "failed to set upload bandwidth control")
	}

	// TODO: Don't start the ticker until the genesis time has been reached and the desired fork has
	// been activated.
	reductionCount := uint(0)
	ticker := time.NewTicker(slotDuration)
	defer ticker.Stop()

	for {
		<-ticker.C
		elapsedEpochs := t.getElapsedEpochs()

		// Reduce bandwidth every two epochs
		if reductionCount < elapsedEpochs/2 {
			// Run the checks.
			if err := runner.RunChecks(context.Background()); err != nil {
				log.Error("Failed to run checks", "error", err)
			} else {
				log.Info("Check results", "results", runner.GetResults())
				log.Info("Check analysis", "analysis", runner.GetAnalysis())
				// log.Info("Check logs", "logs", runner.GetLog().GetBuffer().String())
			}

			reduction := t.currentBandwidth * t.cfg.delta / 100
			if t.currentBandwidth-reduction < t.cfg.minBandwidth {
				log.Info("Bandwidth dropped below minimum threshold, stopping test", "final_bandwidth", FormatBandwidth(t.currentBandwidth), "min_bandwidth", FormatBandwidth(t.cfg.minBandwidth))
				doneChannel <- struct{}{}
				return nil
			}

			t.currentBandwidth -= reduction

			if err := UpdateUploadBandwidthControl(service, t.currentBandwidth); err != nil {
				log.Error("Failed to update bandwidth", "error", err, "epoch", elapsedEpochs, "bandwidth", t.currentBandwidth)
				continue
			}

			nextReduction := t.startTime.Add(slotDuration * time.Duration((elapsedEpochs+2)*slotsPerEpoch))
			log.Info("Reduced bandwidth", "epochs", elapsedEpochs, "new_bandwidth", FormatBandwidth(t.currentBandwidth), "next_reduction_at", nextReduction.Local().Format("15:04:05"))
			reductionCount++
		}
	}
}
