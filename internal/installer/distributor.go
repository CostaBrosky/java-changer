package installer

// Distributor represents a Java distribution provider
type Distributor interface {
	Name() string
	GetAvailableVersions() ([]JavaRelease, error)
	GetDownloadURL(version string, arch string) (*DownloadInfo, error)
}

// JavaRelease represents an available Java version
type JavaRelease struct {
	Version        string
	IsLTS          bool
	OpenJDKVersion string
}

// DownloadInfo contains information needed to download a JDK
type DownloadInfo struct {
	URL          string
	Checksum     string
	ChecksumAlgo string
	Size         int64
	FileName     string
}
