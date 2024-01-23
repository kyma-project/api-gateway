package version

var (
	// dev is the fallback version, as the variable is set to the correct version by the Dockerfile during build time.
	version string = "dev"
)

func GetModuleVersion() string {
	return version
}
