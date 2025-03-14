package tester

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
)

func GetEnclaveContext(ctx context.Context, name string) (*enclaves.EnclaveContext, error) {
	kctx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, err
	}

	enclave, err := kctx.GetEnclaveContext(ctx, name)
	if err != nil {
		return nil, err
	}
	return enclave, nil
}

func CleanupEnclave(enclaveContext *enclaves.EnclaveContext) {

}
