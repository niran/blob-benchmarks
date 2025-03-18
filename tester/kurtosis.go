package tester

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/pkg/errors"
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

func GetOnlyEnclaveContext(ctx context.Context) (*enclaves.EnclaveContext, error) {
	kctx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kurtosis context")
	}

	enclaves, err := kctx.GetEnclaves(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get enclaves")
	}

	enclavesByName := enclaves.GetEnclavesByName()
	if len(enclavesByName) != 1 {
		return nil, fmt.Errorf("expected 1 enclave, got %d", len(enclavesByName))
	}

	var enclaveInfos []*kurtosis_engine_rpc_api_bindings.EnclaveInfo
	for _, enclaveInfos = range enclavesByName {
		break
	}

	if len(enclaveInfos) != 1 {
		return nil, fmt.Errorf("expected 1 enclave info, got %d", len(enclaveInfos))
	}

	enclaveContext, err := kctx.GetEnclaveContext(ctx, enclaveInfos[0].EnclaveUuid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get enclave context")
	}

	return enclaveContext, nil
}

func CleanupEnclave(enclaveContext *enclaves.EnclaveContext) {

}

func GetServiceUnderTest(enclaveContext *enclaves.EnclaveContext) (*services.ServiceContext, error) {
	service, err := enclaveContext.GetServiceContext("cl-1-prysm-geth")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service context")
	}
	log.Info("Retrieved service context", "name", service.GetServiceName(), "uuid", service.GetServiceUUID())
	return service, nil
}
