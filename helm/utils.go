package helm

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func newConfigFlags(config *rest.Config, namespace string) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		Namespace:   &namespace,
		APIServer:   &config.Host,
		CAFile:      &config.CAFile,
		BearerToken: &config.BearerToken,
	}
}

func newActionConf(kubeconf *rest.Config, namespace string, logger *logrus.Entry) (*action.Configuration, error) {
	logger.Infoln("Creating action config ...")
	clientConfig := newConfigFlags(kubeconf, namespace)

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(clientConfig, namespace, "", logger.Infof); err != nil {
		return nil, err
	}

	registryClient, err := registry.NewClient(
		registry.ClientOptWriter(os.Stdout),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)

	if err != nil {
		return nil, err
	}

	actionConfig.RegistryClient = registryClient

	logger.Infoln("Successfully created action config ...")

	return actionConfig, nil
}

func prepareChart(chartPathOption *action.ChartPathOptions, chart string, updateDependency bool, logger *logrus.Entry) (*chart.Chart, error) {
	// Prepare chart
	chartPath, err := chartPathOption.LocateChart(chart, settings)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorf("Failed to locate chart: %s\n", chart)
		return nil, err
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorf("Failed to load chart %s from %s.\n", chart, chartPath)
		return nil, err
	}
	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Warnf("Chart: %s is deprecated.\n", chart)
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if updateDependency {
				man := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        chartPath,
					Keyring:          chartPathOption.Keyring,
					SkipUpdate:       false,
					Getters:          getter.All(settings),
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(chartPath); err != nil {
					logger.WithFields(logrus.Fields{
						"Reason": err.Error(),
					}).Errorf("Failed to load chart %s from %s.\n", chart, chartPath)
					return nil, err
				}
			} else {
				logger.WithFields(logrus.Fields{
					"Reason": err.Error(),
				}).Errorf("Failed to check dependencies of chart %s.\n", chart)
				return nil, err
			}
		}
	}
	return chartRequested, nil

}
