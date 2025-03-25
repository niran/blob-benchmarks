package checks

import (
	"net/http"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethpandaops/panda-pulse/pkg/checks"
	"github.com/ethpandaops/panda-pulse/pkg/clients"
	"github.com/ethpandaops/panda-pulse/pkg/grafana"
)

func SetupRunner(grafanaBaseURL string, grafanaToken string, datasourceID string) (checks.Runner, error) {
	runner := checks.NewDefaultRunner(checks.Config{
		Network: "kurtosis",
		// RunChecks
		ConsensusNode: clients.CLPrysm,
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
	// CLFinalizedEpochCheck's query breaks when joining on `network`.
	// runner.RegisterCheck(checks.NewCLFinalizedEpochCheck(grafanaClient))
	runner.RegisterCheck(checks.NewELSyncCheck(grafanaClient))
	runner.RegisterCheck(checks.NewELBlockHeightCheck(grafanaClient))
	runner.RegisterCheck(NewFailedAttestationsCheck(grafanaClient))
	runner.RegisterCheck(NewFailedProposalsCheck(grafanaClient))

	return runner, nil
}
