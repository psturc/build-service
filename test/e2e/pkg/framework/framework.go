package framework

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/has"
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/imagecontroller"
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/integration"
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/kube"
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/release"
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/tekton"
	"github.com/konflux-ci/build-service/test/e2e/pkg/constants"
	"github.com/konflux-ci/build-service/test/e2e/pkg/logs"
	ginkgo "github.com/onsi/ginkgo/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type ControllerHub struct {
	HasController         *has.Controller
	CommonController      *kube.Controller
	TektonController      *tekton.Controller
	ReleaseController     *release.Controller
	IntegrationController *integration.Controller
	ImageController       *imagecontroller.ImageController
}

type Framework struct {
	AsKubeAdmin          *ControllerHub
	ClusterAppDomain     string
	OpenshiftConsoleHost string
	UserNamespace        string
}

func NewFramework(userName string) (*Framework, error) {
	return NewFrameworkWithTimeout(userName, time.Second*60)
}

func NewFrameworkWithTimeout(userName string, timeout time.Duration) (*Framework, error) {
	if userName == "" {
		return nil, fmt.Errorf("userName cannot be empty when initializing a new framework instance")
	}

	client, err := kube.NewAdminKubernetesClient()
	if err != nil {
		return nil, err
	}

	asAdmin, err := InitControllerHub(client)
	if err != nil {
		return nil, fmt.Errorf("error when initializing appstudio hub controllers for admin user: %v", err)
	}

	nsName := os.Getenv(constants.E2E_APPLICATIONS_NAMESPACE_ENV)
	if nsName == "" {
		nsName = userName
		_, err := asAdmin.CommonController.CreateTestNamespace(userName)
		if err != nil {
			return nil, fmt.Errorf("failed to create test namespace %s: %+v", nsName, err)
		}
	}

	var clusterAppDomain string
	if os.Getenv(constants.TEST_ENVIRONMENT_ENV) == constants.UpstreamTestEnvironment {
		kubeconfig, err := config.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("error when getting kubeconfig: %+v", err)
		}
		parsedURL, err := url.Parse(kubeconfig.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubeconfig host URL: %+v", err)
		}
		clusterAppDomain = parsedURL.Hostname()
	}

	return &Framework{
		AsKubeAdmin:          asAdmin,
		ClusterAppDomain:     clusterAppDomain,
		OpenshiftConsoleHost: clusterAppDomain,
		UserNamespace:        nsName,
	}, nil
}

func InitControllerHub(cc *kube.CustomClient) (*ControllerHub, error) {
	commonCtrl, err := kube.NewController(cc)
	if err != nil {
		return nil, err
	}

	hasController, err := has.NewController(cc)
	if err != nil {
		return nil, err
	}

	tektonController := tekton.NewController(cc)

	releaseController, err := release.NewController(cc)
	if err != nil {
		return nil, err
	}

	integrationController, err := integration.NewController(cc)
	if err != nil {
		return nil, err
	}

	imageController, err := imagecontroller.NewController(cc)
	if err != nil {
		return nil, err
	}

	return &ControllerHub{
		HasController:         hasController,
		CommonController:      commonCtrl,
		TektonController:      tektonController,
		ReleaseController:     releaseController,
		IntegrationController: integrationController,
		ImageController:       imageController,
	}, nil
}

func BuildSuiteDescribe(text string, args ...interface{}) bool {
	return ginkgo.Describe("[build-service-suite "+text+"]", args...)
}

func ReportFailure(f **Framework) func() {
	namespaces := map[string]string{
		"Build Service":       "build-service",
		"Application Service": "application-service",
		"Image Controller":    "image-controller",
	}

	return func() {
		if !ginkgo.CurrentSpecReport().Failed() {
			return
		}

		fwk := *f
		if fwk == nil {
			return
		}

		if err := logs.StoreTestTiming(); err != nil {
			ginkgo.GinkgoWriter.Printf("failed to store test timing: %v\n", err)
		}

		allPodLogs := make(map[string][]byte)
		for _, namespace := range namespaces {
			podList, err := fwk.AsKubeAdmin.CommonController.ListAllPods(namespace)
			if err != nil {
				ginkgo.GinkgoWriter.Printf("failed to list pods in namespace %s: %v\n", namespace, err)
				continue
			}

			for _, pod := range podList.Items {
				podLogs := fwk.AsKubeAdmin.CommonController.GetPodLogs(&pod)

				for podName, log := range podLogs {
					if filteredLogs := FilterLogs(string(log), ginkgo.CurrentSpecReport().StartTime); filteredLogs != "" {
						allPodLogs[podName] = []byte(filteredLogs)
					}
				}
			}
		}

		if err := logs.StoreArtifacts(allPodLogs); err != nil {
			ginkgo.GinkgoWriter.Printf("failed to store pod logs: %v\n", err)
		}
	}
}

func FilterLogs(logStr string, start time.Time) string {
	lines := strings.Split(logStr, "\n")
	var ret []string
	rfc3339Pattern := `(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2}))`

	re := regexp.MustCompile(rfc3339Pattern)
	for pos, line := range lines {
		match := re.FindStringSubmatch(line)
		if match != nil {
			dateString := match[1]
			ts, err := time.Parse(time.RFC3339, dateString)
			if err != nil {
				ret = append(ret, "Invalid Time, unable to parse date: "+line)
			} else if ts.Equal(start) || ts.After(start) {
				ret = append(ret, lines[pos:]...)
				break
			}
		}
	}

	return strings.Join(ret, "\n")
}
