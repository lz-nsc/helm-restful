package helm

import (
	"helm-restful/helm/api"
	"time"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/rest"
)

func retrieve(kubeconf *rest.Config, namespace string, name string, logger *logrus.Entry) (*release.Release, error) {
	logger.Infof("Start to get info of releases %s from namesapce %s\n", name, namespace)

	actionConfig, err := newActionConf(kubeconf, namespace, logger)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorln("Failed to initialize action config.")
		return nil, err
	}

	client := action.NewGet(actionConfig)
	res, err := client.Run(name)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorf("Failed to get info of releases %s from namesapce %s\n", name, namespace)
		return nil, err
	}
	return res, nil
}

func list(kubeconf *rest.Config, namespace string, logger *logrus.Entry) ([]*release.Release, error) {
	logger.Infof("Start to get info of all releases from namesapce %s\n", namespace)
	actionConfig, err := newActionConf(kubeconf, namespace, logger)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorln("Failed to initialize action config.")
		return nil, err
	}

	client := action.NewList(actionConfig)
	client.SetStateMask()
	return client.Run()
}

func install(kubeconf *rest.Config, helmData *api.HelmData, logger *logrus.Entry) (*release.Release, error) {

	logger.Infoln("Installing helm chart ...")

	namespace := helmData.GetNamespace()
	name := helmData.Name

	// Initialize actionConfig
	actionConfig, err := newActionConf(kubeconf, namespace, logger)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorln("Failed to initialize action config.")
		return nil, err
	}
	chart := helmData.Chart

	client := action.NewInstall(actionConfig)
	client.ReleaseName = name
	client.CreateNamespace = true
	client.Namespace = namespace
	client.Timeout = 300 * time.Second

	// Prepare chart
	chartRequested, err := prepareChart(&client.ChartPathOptions, chart, client.DependencyUpdate, logger)
	if err != nil {
		return nil, err
	}

	return client.Run(chartRequested, helmData.Values)

}

func update(kubeconf *rest.Config, namespace string, name string, helmUpdateData *api.HelmUpdateData, logger *logrus.Entry) (*release.Release, error) {
	logger.Infof("Start to update release %s from namesapce %s\n", name, namespace)
	actionConfig, err := newActionConf(kubeconf, namespace, logger)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorln("Failed to initialize action config.")
		return nil, err
	}
	chart := helmUpdateData.Chart

	client := action.NewUpgrade(actionConfig)
	chartRequested, err := prepareChart(&client.ChartPathOptions, chart, client.DependencyUpdate, logger)
	if err != nil {
		return nil, err
	}

	res, err := client.Run(name, chartRequested, helmUpdateData.Values)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorf("Failed to update releases %s from namesapce %s\n", name, namespace)
		return nil, err
	}
	return res, nil
}

func delete(kubeconf *rest.Config, namespace string, name string, logger *logrus.Entry) error {
	logger.Infof("Start to delete release %s from namesapce %s\n", name, namespace)
	actionConfig, err := newActionConf(kubeconf, namespace, logger)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorln("Failed to initialize action config.")
		return err
	}

	client := action.NewUninstall(actionConfig)
	_, err = client.Run(name)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Reason": err.Error(),
		}).Errorf("Failed to release %s from namesapce %s\n", name, namespace)
		return err
	}

	logger.Infof("Successfully delete release %s from namesapce %s\n", name, namespace)

	return nil
}
