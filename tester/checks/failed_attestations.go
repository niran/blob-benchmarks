package checks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethpandaops/panda-pulse/pkg/checks"
	"github.com/ethpandaops/panda-pulse/pkg/clients"
	"github.com/ethpandaops/panda-pulse/pkg/grafana"
	"github.com/ethpandaops/panda-pulse/pkg/logger"
)

const queryFailedAttestations = `
	validator_failed_attestations{network=~"%s", client_name=~"%s"} > 0
`

// FailedAttestationsCheck is a check that verifies if validators are failing attestations.
type FailedAttestationsCheck struct {
	grafanaClient grafana.Client
}

// NewFailedAttestationsCheck creates a new FailedAttestationsCheck.
func NewFailedAttestationsCheck(grafanaClient grafana.Client) *FailedAttestationsCheck {
	return &FailedAttestationsCheck{
		grafanaClient: grafanaClient,
	}
}

// Name returns the name of the check.
func (c *FailedAttestationsCheck) Name() string {
	return "Validators failing attestations"
}

// Category returns the category of the check.
func (c *FailedAttestationsCheck) Category() checks.Category {
	return checks.CategoryGeneral
}

// ClientType returns the client type of the check.
func (c *FailedAttestationsCheck) ClientType() clients.ClientType {
	return clients.ClientTypeCL
}

// Run executes the check.
func (c *FailedAttestationsCheck) Run(ctx context.Context, log *logger.CheckLogger, cfg checks.Config) (*checks.Result, error) {
	query := fmt.Sprintf(queryFailedAttestations, cfg.Network, cfg.ConsensusNode)

	log.Print("\n=== Running failed attestations check")

	response, err := c.grafanaClient.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Track validators with failed attestations
	type ValidatorInfo struct {
		Instance string
		PubKey   string
	}
	var failingValidators []ValidatorInfo

	for _, frame := range response.Results.PandaPulse.Frames {
		for _, field := range frame.Schema.Fields {
			if labels := field.Labels; labels != nil {
				if labels["instance"] != "" && labels["pubkey"] != "" {
					validator := ValidatorInfo{
						Instance: labels["instance"],
						PubKey:   labels["pubkey"],
					}
					failingValidators = append(failingValidators, validator)
					log.Printf("  - Failing validator: %s on instance %s", validator.PubKey, validator.Instance)
				}
			}
		}
	}

	if len(failingValidators) == 0 {
		log.Printf("  - No validators are failing attestations")

		return &checks.Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      checks.StatusOK,
			Description: "All validators are attesting properly",
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"query": query,
			},
			AffectedNodes: []string{},
		}, nil
	}

	// Create list of affected nodes (instances)
	affectedNodes := make([]string, 0)
	nodeSet := make(map[string]bool)
	var validatorDetails []string

	for _, v := range failingValidators {
		if !nodeSet[v.Instance] {
			affectedNodes = append(affectedNodes, v.Instance)
			nodeSet[v.Instance] = true
		}
		validatorDetails = append(validatorDetails, fmt.Sprintf("Instance: %s, PubKey: %s", v.Instance, v.PubKey))
	}

	return &checks.Result{
		Name:        c.Name(),
		Category:    c.Category(),
		Status:      checks.StatusFail,
		Description: "Some validators are failing attestations",
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"query":             query,
			"failingValidators": strings.Join(validatorDetails, "\n"),
		},
		AffectedNodes: affectedNodes,
	}, nil
}
