package tester

import "github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"

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

func (t *MinBandwidthTest) Run(doneChannel chan struct{}) error {
	// TODO: Get the service for the node whose bandwidth we want to limit.

	// TODO: Set the bandwidth limit for the service.

	// TODO: Run the test.

	doneChannel <- struct{}{}
	return nil
}
