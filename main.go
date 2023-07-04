package main

import (
	"fmt"
	"helm-restful/helm"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceQuote:      true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	var (
		kubeConfigPath string
		port           string
		host           string
	)
	pflag.StringVar(&kubeConfigPath, "kubeconfig", os.Getenv("KUBECONFIG"), "path to kubeconfig file of target cluster")
	pflag.StringVarP(&port, "port", "p", "8080", "port to which the server will listen")
	pflag.StringVarP(&host, "host", "h", "0.0.0.0", "host to which the server will listen")

	helm.AddHelmFlags()

	pflag.Parse()

	kubeconf, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath}, nil).ClientConfig()

	if err != nil {
		logrus.Fatalf("Failed to initialize config with given kubeConfig: %s.", kubeConfigPath)
	}

	helm.InitializeHelmWebService(kubeconf)
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	logrus.Infof("Start to listen to %s ...", serverAddr)
	logrus.Fatal(http.ListenAndServe(serverAddr, nil))
}
