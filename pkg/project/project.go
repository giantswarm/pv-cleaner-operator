package project

var (
	description string = "The pv-cleaner-operator cleans up released k8s persistent volumes."
	gitSHA             = "n/a"
	name        string = "pv-cleaner-operator"
	source      string = "https://github.com/giantswarm/pv-cleaner-operator"
	version            = "n/a"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
