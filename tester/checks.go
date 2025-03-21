package tester

import (
	"net/http"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethpandaops/panda-pulse/pkg/checks"
	"github.com/ethpandaops/panda-pulse/pkg/clients"
	"github.com/ethpandaops/panda-pulse/pkg/grafana"
)

func SetupRunner(grafanaBaseURL string, grafanaToken string, datasourceID string) (checks.Runner, error) {
	runner := checks.NewDefaultRunner(checks.Config{
		Network:       "kurtosis",
		ConsensusNode: clients.CLPrysm,
		ExecutionNode: clients.ELGeth,
	})

	httpClient := &http.Client{}
	grafanaClient := grafana.NewClient(&grafana.Config{
		BaseURL:          grafanaBaseURL,
		Token:            grafanaToken,
		PromDatasourceID: datasourceID,
	}, httpClient)

	log.Debug("Grafana client", "client", grafanaClient)

	runner.RegisterCheck(checks.NewCLSyncCheck(grafanaClient))
	runner.RegisterCheck(checks.NewHeadSlotCheck(grafanaClient))
	runner.RegisterCheck(checks.NewCLFinalizedEpochCheck(grafanaClient))
	runner.RegisterCheck(checks.NewELSyncCheck(grafanaClient))
	runner.RegisterCheck(checks.NewELBlockHeightCheck(grafanaClient))

	return runner, nil
}
