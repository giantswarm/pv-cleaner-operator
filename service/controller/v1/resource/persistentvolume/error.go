package persistentvolume

import (
	"github.com/giantswarm/microerror"
	"strings"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}

// IsPVCExists asserts pvc exists error from create API response.
func IsPVCExists(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(microerror.Cause(err).Error(), "already exists")
}
