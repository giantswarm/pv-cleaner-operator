package service

import (
	"github.com/giantswarm/pv-cleaner-operator/flag/service/kubernetes"
)

type Service struct {
	Kubernetes kubernetes.Kubernetes
}
