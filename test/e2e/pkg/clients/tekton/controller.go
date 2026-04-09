package tekton

import (
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/kube"
)

// Create the struct for kubernetes clients
type Controller struct {
	*kube.CustomClient
}

// Create controller for Tekton Task/Pipeline CRUD operations
func NewController(kube *kube.CustomClient) *Controller {
	return &Controller{
		kube,
	}
}
