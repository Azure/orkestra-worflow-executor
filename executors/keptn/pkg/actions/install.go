package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/orkestra-workflow-executor/executors/keptn/pkg/keptn"
	fluxhelmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
)

func Install(ctx context.Context, cancel context.CancelFunc, hr *fluxhelmv2beta1.HelmRelease, interval time.Duration, data map[string]string) error {
	keptnConfig := &keptn.KeptnConfig{}

	// Read the keptn-config.yaml file.
	// This file is required and cannot have empty fields
	if v, ok := data[keptn.KeptnConfigFileName]; !ok {
		return fmt.Errorf("failed to read plugin configuration from configmap")
	} else {
		if err := json.Unmarshal([]byte(v), keptnConfig); err != nil {
			return fmt.Errorf("failed to unmarshal the keptn configuration file into KeptnConfig object")
		}
	}

	if err := keptnConfig.Validate(); err != nil {
		return err
	}

	keptnCli, err := keptn.New(keptnConfig.URL, keptnConfig.Namespace, keptnConfig.Token.SecretRef.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to create the keptn client %w", err)
	}

	shipyard, ok := data[keptn.ShipyardFileName]
	if !ok {
		return fmt.Errorf("shipyard.yaml not found")
	}

	appName := strings.ToLower(hr.Name + "-" + hr.Namespace)
	if err := keptnCli.CreateProject(appName, shipyard); err != nil {
		// if err := keptnCli.CreateProject(strings.ToLower("new-evaluation-project"), []byte(shipyard)); err != nil {
		return err
	}

	if err := keptnCli.CreateService(appName, appName); err != nil {
		return err
	}

	for k, v := range data {
		if err := keptnCli.AddResourceToAllStages(appName, appName, k, v); err != nil {
			return err
		}
	}

	if err := keptnCli.ConfigureMonitoring(appName, appName, "prometheus"); err != nil {
		return err
	}

	if err := keptnCli.TriggerEvaluation(appName, appName, keptnConfig.Timeframe); err != nil {
		return err
	}
	return nil
}
