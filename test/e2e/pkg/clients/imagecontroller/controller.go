package imagecontroller

import (
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/kube"
)

type ImageController struct {
	*kube.CustomClient
}

func NewController(kube *kube.CustomClient) (*ImageController, error) {
	return &ImageController{
		kube,
	}, nil
}
