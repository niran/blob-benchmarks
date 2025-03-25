package tester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/pkg/errors"
)

type GrafanaDatasource struct {
	ID   int    `json:"id"`
	UID  string `json:"uid"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func GetGrafanaBaseURL(enclaveContext *enclaves.EnclaveContext) (string, error) {
	service, err := enclaveContext.GetServiceContext("grafana")
	if err != nil {
		return "", errors.Wrap(err, "failed to get grafana service context")
	}

	ipAddress := service.GetMaybePublicIPAddress()
	ports := service.GetPublicPorts()

	var httpPort *services.PortSpec
	for _, port := range ports {
		httpPort = port
		break
	}

	if httpPort == nil {
		return "", fmt.Errorf("http port not found")
	}

	return fmt.Sprintf("http://%s:%d", ipAddress, httpPort.GetNumber()), nil
}

func GetGrafanaConfig(enclaveContext *enclaves.EnclaveContext) (string, string, string, error) {
	grafanaBaseURL, err := GetGrafanaBaseURL(enclaveContext)
	if err != nil {
		return "", "", "", errors.Wrap(err, "failed to get grafana base URL")
	}

	apiToken, isSet := os.LookupEnv("GRAFANA_API_TOKEN")
	if !isSet {
		apiToken, err = createGrafanaServiceAccountAndToken(grafanaBaseURL)
		if err != nil {
			return "", "", "", errors.Wrap(err, "failed to create service account and token")
		}
		log.Info("Created service account", "token", apiToken)
	}

	datasourceID, isSet := os.LookupEnv("GRAFANA_DATASOURCE_ID")
	if !isSet {
		datasourceID, err = GetGrafanaDatasourceID(grafanaBaseURL, apiToken)
		if err != nil {
			return "", "", "", errors.Wrap(err, "failed to get datasource ID")
		}
		log.Info("Retrieved datasource ID", "id", datasourceID)
	}

	return grafanaBaseURL, apiToken, datasourceID, nil
}

// Query the API to get the datasource ID
func GetGrafanaDatasourceID(grafanaBaseURL string, apiToken string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", grafanaBaseURL+"/api/datasources", nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create request for datasources")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	log.Info("Getting datasources", "url", grafanaBaseURL+"/api/datasources")
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to get datasources")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get datasources: status %d, body: %s", resp.StatusCode, string(body))
	}

	var datasources []GrafanaDatasource
	if err := json.NewDecoder(resp.Body).Decode(&datasources); err != nil {
		return "", errors.Wrap(err, "failed to decode datasources response")
	}

	if len(datasources) == 0 {
		return "", fmt.Errorf("no datasources found")
	}

	// Use the first datasource's UID
	datasourceID := datasources[0].UID
	return datasourceID, nil
}

// Create a service account and return its token
func createGrafanaServiceAccountAndToken(grafanaBaseURL string) (string, error) {
	// Create service account
	serviceAccountName := "benchmarks"
	client := &http.Client{}
	serviceAccountPayload := map[string]interface{}{
		"name":       serviceAccountName,
		"role":       "Viewer",
		"isDisabled": false,
	}

	jsonPayload, err := json.Marshal(serviceAccountPayload)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal service account payload")
	}

	req, err := http.NewRequest("POST", grafanaBaseURL+"/api/serviceaccounts", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", errors.Wrap(err, "failed to create request for service account creation")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth("admin", "admin")

	log.Info("Creating service account", "name", serviceAccountName)
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to create service account")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create service account: status %d, body: %s", resp.StatusCode, string(body))
	}

	var serviceAccount struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&serviceAccount); err != nil {
		return "", errors.Wrap(err, "failed to decode service account response")
	}

	// Create token for the service account
	tokenPayload := map[string]interface{}{
		"name": serviceAccountName,
	}

	jsonPayload, err = json.Marshal(tokenPayload)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal token payload")
	}

	req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/serviceaccounts/%d/tokens", grafanaBaseURL, serviceAccount.ID), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", errors.Wrap(err, "failed to create request for token creation")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth("admin", "admin")

	log.Info("Creating service account token", "name", serviceAccountName)
	resp, err = client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to create token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create token: status %d, body: %s", resp.StatusCode, string(body))
	}

	var token struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", errors.Wrap(err, "failed to decode token response")
	}

	return token.Key, nil
}
