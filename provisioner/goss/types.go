package goss

const (
	// Install
	Linux       = "Linux"
	Windows     = "Windows"
	DefaultArch = "amd64"
	DefaultOS   = Linux
	Latest      = "latest"

	// LatestVersionDownloadURL <os> <arch>.
	LatestVersionDownloadURL = "https://github.com/goss-org/goss/releases/latest/download/goss-%s-%s"

	// VersionDownloadURL <version> <os> <arch>.
	VersionDownloadURL = "https://github.com/goss-org/goss/releases/download/v%s/goss-%s-%s"

	// DefaultDownloadPath /tmp/goss-<version>-<os>-<arch>.
	DefaultDownloadPath = "/tmp/goss-%s-%s-%s"

	// DownloadCmd <sudo> <env vars> <ssl flag> <basic auth creds> <URL> <output path> || <sudo> <env vars> <ssl flag> <basic auth creds> <URL> <output path>.
	DownloadCmd = "%s %s curl -sL %s %s -o %s %s || %s %s wget -q %s %s -O %s %s"

	// InstallCmd < goss binary path> <goss binary path>.
	InstallCmd = "chmod 555 %s && %s --version"

	// Render
	// <remote path> <env vars <sudo> <goss binary path> <goss file> <goss vars> <goss inline vars> <output file>.
	// RenderCmd = "cd %s && %s %s %s %s %s %s render %s > %s"

	// // DefaultGossFile <remote path>.
	// DefaultGossSpecFile = "./goss-spec.yaml"
	DefaultGossFile = "./goss.yaml"

	// Validate
	// <env vars <sudo> <goss binary path> <loglevel> <package type> <goss file> <goss vars> <goss inline vars> <retry timeout> <sleep> <format> <format options> <output file>.
	ValidateCmd = "%s %s %s %s %s %s %s %s validate --retry-timeout=%s --sleep=%s %s %s %s"

	// mkDirCmd <path>
	mkDirCmd = "mkdir -p %s"

	DefaultSleep        = "1s"
	DefaultRetryTimeout = "0s"

	// We dont expose the RemotePath in the validate block so we avoid confusing user with directories and force them to always use absolute paths
	DefaultRemotePath = "/tmp"
)

var (
	validOS            = []string{Linux, Windows}
	ValidFormats       = []string{"documentation", "json", "json_oneline", "junit", "nagios", "nagios_verbose", "rspecish", "silent", "tap"}
	ValidFormatOptions = []string{"perfdata", "verbose", "pretty"}
	ValidPackageTypes  = []string{"apk", "dpkg", "pacman", "rpm"}
	validLogLevel      = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}
)
