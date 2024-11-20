package config

const (
	// VersionSlinky indicates that the chain we are concerned about is using slinky v1.
	VersionSlinky Version = "slinky"

	// VersionConnect indicates that the chain we are concerned about is using connect v2.
	VersionConnect Version = "connect"
)

type Version string

func IsValidVersion(version Version) bool {
	return version == VersionSlinky || version == VersionConnect
}
