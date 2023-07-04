package api

import "helm.sh/helm/v3/pkg/release"

type HelmUpdateData struct {
	Chart  string                 `json:"chart" validate:"required"`
	Values map[string]interface{} `json:"values,omitempty"`
}

type HelmData struct {
	Name      string                 `json:"name" validate:"required"`
	Namespace *string                `json:"namespace,omitempty"`
	Chart     string                 `json:"chart" validate:"required"`
	Values    map[string]interface{} `json:"values,omitempty"`
}

func (data HelmData) GetNamespace() string {
	if data.Namespace == nil {
		return "default"
	}
	return *data.Namespace
}

type HelmReleaseInfo struct {
	Name        string `json:"name"`
	Status      string `json:"status,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Version     int    `json:"version,omitempty"`
	DeployedAt  string `json:"deployed_at"`
	Discription string `json:"description,omitempty"`
}

type HelmReleaseList struct {
	TotalItems int               `json:"total_items"`
	Releases   []HelmReleaseInfo `json:"releases"`
}

func ToHelmReleaseInfo(release *release.Release) HelmReleaseInfo {
	return HelmReleaseInfo{
		Name:        release.Name,
		Status:      string(release.Info.Status),
		Namespace:   release.Namespace,
		Version:     release.Version,
		DeployedAt:  release.Info.FirstDeployed.Format("2006-01-02 15:04:05"),
		Discription: release.Info.Description,
	}
}

func ToHelmReleaseList(releases []*release.Release) HelmReleaseList {
	releaseList := HelmReleaseList{
		TotalItems: len(releases),
		Releases:   make([]HelmReleaseInfo, 0),
	}

	for _, release := range releases {
		info := ToHelmReleaseInfo(release)
		releaseList.Releases = append(releaseList.Releases, info)
	}
	return releaseList
}
