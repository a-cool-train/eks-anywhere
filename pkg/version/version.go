package version

var gitVersion string

type Info struct {
	GitVersion string
}

func Get() Info {
	return Info{
		GitVersion: "v0.0.0-dev",
	}
}
