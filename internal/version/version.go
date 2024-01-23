package version

const defaultModuleVersion string = "dev"

var (
	// variable is set to the correct version by the Dockerfile during build time.
	version string
)

func GetModuleVersion() string {
	if version == "" {
		version = defaultModuleVersion
	}
	return version
}
