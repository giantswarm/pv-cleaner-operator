package persistentvolume

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var wrongTypeError = microerror.New("wrong type")

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}

var missingAnnotationError = microerror.New("missing annotation")

// IsMissingAnnotationError asserts missingAnnotationError.
func IsMissingAnnotationError(err error) bool {
	return microerror.Cause(err) == missingAnnotationError
}

var waitResourceError = microerror.New("waitResource")

// IsWaitTimeout asserts invalidConfigError.
func IsWaitReosurce(err error) bool {
	return microerror.Cause(err) == waitResourceError
}

var waitTimeoutError = microerror.New("waitTimeout")

// IsWaitTimeout asserts invalidConfigError.
func IsWaitTimeout(err error) bool {
	return microerror.Cause(err) == waitTimeoutError
}