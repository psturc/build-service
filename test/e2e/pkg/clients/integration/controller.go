package integration

import (
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/kube"
)

type Controller struct {
	*kube.CustomClient
}

func NewController(kube *kube.CustomClient) (*Controller, error) {
	return &Controller{
		kube,
	}, nil
}
