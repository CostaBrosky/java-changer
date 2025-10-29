package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

const adoptiumAPIBase = "https://api.adoptium.net/v3"

// AdoptiumDistributor implements the Distributor interface for Eclipse Adoptium
type AdoptiumDistributor struct{}

// NewAdoptiumDistributor creates a new Adoptium distributor
func NewAdoptiumDistributor() *AdoptiumDistributor {
	return &AdoptiumDistributor{}
}

// Name returns the distributor name
func (a *AdoptiumDistributor) Name() string {
	return "Eclipse Adoptium"
}

// adoptiumReleasesResponse represents the API response for available releases
type adoptiumReleasesResponse struct {
	AvailableLTSReleases     []int `json:"available_lts_releases"`
	AvailableReleases        []int `json:"available_releases"`
	MostRecentLTS            int   `json:"most_recent_lts"`
	MostRecentFeatureRelease int   `json:"most_recent_feature_release"`
}

// adoptiumAssetResponse represents the API response for asset details
type adoptiumAssetResponse struct {
	Binary struct {
		Package struct {
			Link     string `json:"link"`
			Checksum string `json:"checksum"`
			Size     int64  `json:"size"`
			Name     string `json:"name"`
		} `json:"package"`
	} `json:"binary"`
	Version struct {
		OpenJDKVersion string `json:"openjdk_version"`
		Major          int    `json:"major"`
	} `json:"version"`
}

// GetAvailableVersions fetches available Java versions from Adoptium API
func (a *AdoptiumDistributor) GetAvailableVersions() ([]JavaRelease, error) {
	url := fmt.Sprintf("%s/info/available_releases", adoptiumAPIBase)

	resp, err := http.Get(url)
	if err != nil {
		return a.getFallbackVersions(), fmt.Errorf("API request failed, using fallback versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a.getFallbackVersions(), fmt.Errorf("API returned status %d, using fallback versions", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return a.getFallbackVersions(), fmt.Errorf("failed to read response, using fallback versions: %w", err)
	}

	var releasesResp adoptiumReleasesResponse
	if err := json.Unmarshal(body, &releasesResp); err != nil {
		return a.getFallbackVersions(), fmt.Errorf("failed to parse response, using fallback versions: %w", err)
	}

	// Convert to JavaRelease structs
	releases := make([]JavaRelease, 0, len(releasesResp.AvailableReleases))
	ltsMap := make(map[int]bool)
	for _, v := range releasesResp.AvailableLTSReleases {
		ltsMap[v] = true
	}

	for _, v := range releasesResp.AvailableReleases {
		releases = append(releases, JavaRelease{
			Version: fmt.Sprintf("%d", v),
			IsLTS:   ltsMap[v],
		})
	}

	// Sort descending by version
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Version > releases[j].Version
	})

	return releases, nil
}

// getFallbackVersions returns a hardcoded list of versions as fallback
func (a *AdoptiumDistributor) getFallbackVersions() []JavaRelease {
	return []JavaRelease{
		{Version: "25", IsLTS: true},
		{Version: "24", IsLTS: false},
		{Version: "23", IsLTS: false},
		{Version: "22", IsLTS: false},
		{Version: "21", IsLTS: true},
		{Version: "20", IsLTS: false},
		{Version: "19", IsLTS: false},
		{Version: "18", IsLTS: false},
		{Version: "17", IsLTS: true},
		{Version: "16", IsLTS: false},
		{Version: "11", IsLTS: true},
		{Version: "8", IsLTS: true},
	}
}

// GetDownloadURL fetches download information for a specific version and architecture
func (a *AdoptiumDistributor) GetDownloadURL(version string, arch string) (*DownloadInfo, error) {
	// Map Go arch to Adoptium arch
	adoptiumArch := arch
	if arch == "amd64" {
		adoptiumArch = "x64"
	}

	url := fmt.Sprintf("%s/assets/latest/%s/hotspot?architecture=%s&image_type=jdk&os=windows&vendor=eclipse",
		adoptiumAPIBase, version, adoptiumArch)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query download URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d for version %s", resp.StatusCode, version)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var assets []adoptiumAssetResponse
	if err := json.Unmarshal(body, &assets); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(assets) == 0 {
		return nil, fmt.Errorf("no JDK found for Java %s on %s", version, arch)
	}

	asset := assets[0]
	return &DownloadInfo{
		URL:          asset.Binary.Package.Link,
		Checksum:     asset.Binary.Package.Checksum,
		ChecksumAlgo: "SHA256",
		Size:         asset.Binary.Package.Size,
		FileName:     asset.Binary.Package.Name,
	}, nil
}
