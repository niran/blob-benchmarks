package tester

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/pkg/errors"
)

type MinBandwidthTestConfig struct {
	enclaveContext *enclaves.EnclaveContext
	blobsPerBlock  uint
	bandwidth      uint
	delta          uint
}

type MinBandwidthTest struct {
	cfg              MinBandwidthTestConfig
	currentBandwidth uint
}

func NewMinBandwidthTest(enclaveContext *enclaves.EnclaveContext, blobsPerBlock uint, bandwidth uint, delta uint) *MinBandwidthTest {
	return &MinBandwidthTest{
		cfg: MinBandwidthTestConfig{
			enclaveContext: enclaveContext,
			blobsPerBlock:  blobsPerBlock,
			bandwidth:      bandwidth,
			delta:          delta,
		},
		currentBandwidth: bandwidth,
	}
}

func NewMinBandwidthTestForOnlyEnclave(ctx context.Context, blobsPerBlock uint, bandwidth uint, delta uint) (*MinBandwidthTest, error) {
	enclaveContext, err := GetOnlyEnclaveContext(ctx)
	if err != nil {
		return nil, err
	}

	return NewMinBandwidthTest(enclaveContext, blobsPerBlock, bandwidth, delta), nil
}

func (t *MinBandwidthTest) Run(doneChannel chan struct{}) error {
	// Get the service for the node whose bandwidth we want to limit.
	service, err := t.cfg.enclaveContext.GetServiceContext("cl-1-prysm-geth")
	if err != nil {
		return errors.Wrap(err, "failed to get service context")
	}
	log.Info("Retrieved service context", "name", service.GetServiceName(), "uuid", service.GetServiceUUID())

	// Install the tc command.
	if err := InstallTcCommand(service); err != nil {
		return errors.Wrap(err, "failed to install tc command")
	}

	// Remove any existing bandwidth controls.
	if err := RemoveBandwidthControls(service); err != nil {
		log.Info("No existing bandwidth controls seem to be set, continuing...", "message", err)
	}

	defer RemoveBandwidthControls(service)

	// Set the download bandwidth to a reasonable fixed value.
	if err := SetDownloadBandwidthControl(service, "100mbit"); err != nil {
		return errors.Wrap(err, "failed to set download bandwidth control")
	}

	// Set the upload bandwith to a starting point for the tests.
	if err := SetUploadBandwidthControl(service, "50mbit"); err != nil {
		return errors.Wrap(err, "failed to set upload bandwidth control")
	}

	// TODO: Run the test.

	doneChannel <- struct{}{}
	return nil
}
