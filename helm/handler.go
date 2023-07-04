package helm

import (
	"helm-restful/helm/api"
	"net/http"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/rest"
)

var baseLog = logrus.WithFields(logrus.Fields{
	"Resource": "Helm",
})
var settings = cli.New()

func AddHelmFlags() {
	pflag.StringVar(&settings.RegistryConfig, "registry-config", settings.RegistryConfig, "path to the registry config file")
	pflag.StringVar(&settings.RepositoryConfig, "repository-config", settings.RepositoryConfig, "path to the file containing repository names and URLs")
	pflag.StringVar(&settings.RepositoryCache, "repository-cache", settings.RepositoryCache, "path to the file containing cached repository indexes")
}

func handleHelmRetrieve(kubeconf *rest.Config) restful.RouteFunction {
	// Set install logger
	retrieveLog := baseLog.WithFields(logrus.Fields{
		"Action": "Retrieve",
	})

	return func(request *restful.Request, response *restful.Response) {
		namespace := request.PathParameter("namespace")
		name := request.PathParameter("name")
		release, err := retrieve(kubeconf, namespace, name, retrieveLog)

		if err != nil {
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusInternalServerError, err.Error()+"\n")
			return
		}

		response.WriteHeaderAndEntity(http.StatusOK, api.ToHelmReleaseInfo(release))
	}
}
func handleHelmList(kubeconf *rest.Config) restful.RouteFunction {
	// Set install logger
	listLog := baseLog.WithFields(logrus.Fields{
		"Action": "List",
	})

	return func(request *restful.Request, response *restful.Response) {
		namespace := request.PathParameter("namespace")
		releases, err := list(kubeconf, namespace, listLog)

		if err != nil {
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusInternalServerError, err.Error()+"\n")
			if namespace != "" {
				listLog.Errorf("Failed to list release from namespace %s", namespace)
				return
			}
			listLog.Errorln("Failed to list release from all namespaces")
			return
		}

		response.WriteHeaderAndEntity(http.StatusOK, api.ToHelmReleaseList(releases))
	}
}

func handleHelmInstall(kubeconf *rest.Config) restful.RouteFunction {
	// Set install logger
	installLog := baseLog.WithFields(logrus.Fields{
		"Action": "Install",
	})

	return func(request *restful.Request, response *restful.Response) {
		helmData := new(api.HelmData)
		if err := request.ReadEntity(helmData); err != nil {
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusInternalServerError, err.Error()+"\n")
		}
		releaseName := helmData.Name
		release, err := install(kubeconf, helmData, installLog)
		if err != nil {
			installLog.WithFields(logrus.Fields{
				"Reason": err.Error(),
			}).Errorf("Failed to intall release %s", releaseName)
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusInternalServerError, err.Error()+"\n")
			return
		}

		installLog.Infof("Successfully installed release %s\n", releaseName)
		response.WriteHeaderAndEntity(http.StatusCreated, api.ToHelmReleaseInfo(release))
	}
}

func handleHelmDelete(kubeconf *rest.Config) restful.RouteFunction {
	// Set install logger
	deleteLog := baseLog.WithFields(logrus.Fields{
		"Action": "Delete",
	})

	return func(request *restful.Request, response *restful.Response) {
		namespace := request.PathParameter("namespace")
		name := request.PathParameter("name")
		err := delete(kubeconf, namespace, name, deleteLog)
		if err != nil {
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusInternalServerError, err.Error()+"\n")
			return
		}

		response.WriteHeader(http.StatusNoContent)
	}
}

func handleHelmUpdate(kubeconf *rest.Config) restful.RouteFunction {
	// Set install logger
	updateLog := baseLog.WithFields(logrus.Fields{
		"Action": "Update",
	})

	return func(request *restful.Request, response *restful.Response) {
		namespace := request.PathParameter("namespace")
		name := request.PathParameter("name")
		helmUpdateData := new(api.HelmUpdateData)

		if err := request.ReadEntity(helmUpdateData); err != nil {
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusInternalServerError, err.Error()+"\n")
		}

		release, err := update(kubeconf, namespace, name, helmUpdateData, updateLog)
		if err != nil {
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusInternalServerError, err.Error()+"\n")
			return
		}

		response.WriteHeaderAndEntity(http.StatusOK, api.ToHelmReleaseInfo(release))
	}
}

// ----------- Install -----------

// ---------- WebService ----------
func InitializeHelmWebService(kubeconf *rest.Config) {
	ws := new(restful.WebService)
	// ws.Path("/helms").
	// 	Consumes(restful.MIME_XML, restful.MIME_JSON).
	// 	Produces(restful.MIME_JSON, restful.MIME_XML)
	// List all helm releases from all namespaces
	ws.Route(
		ws.GET("").
			To(handleHelmList(kubeconf)))
	// List all helm releases from a specified namespace
	ws.Route(
		ws.GET("/{namespace}").
			To(handleHelmList(kubeconf)))
	// Get info of target helm release from a specified namespace
	ws.Route(
		ws.GET("/{namespace}/{name}").
			To(handleHelmRetrieve(kubeconf)))
	// Install helm chart
	ws.Route(
		ws.POST("").
			Reads(api.HelmData{}).
			To(handleHelmInstall(kubeconf)).
			Writes(api.HelmReleaseInfo{}))
	// Upgrate target helm chart
	ws.Route(
		ws.PUT("/{namespace}/{name}").
			Reads(api.HelmData{}).
			To(handleHelmUpdate(kubeconf)))
	// Delete target release from specific namespaces
	ws.Route(
		ws.DELETE("/{namespace}/{name}").
			To(handleHelmDelete(kubeconf)))
	restful.Add(ws)
}
