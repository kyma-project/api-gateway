package version

var (
	version   string = "dev"
	buildTime string
)

func GetVersion() string {
	return version
}

func GetBuildTime() string {
	return buildTime
}
