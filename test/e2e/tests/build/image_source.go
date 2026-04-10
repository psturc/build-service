package build

import (
	"fmt"
	"strings"
	"time"

	"github.com/devfile/library/v2/pkg/util"
	appservice "github.com/konflux-ci/application-api/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck
	. "github.com/onsi/gomega"    //nolint:staticcheck

	"github.com/konflux-ci/build-service/test/e2e/pkg/constants"
	"github.com/konflux-ci/build-service/test/e2e/pkg/framework"
	"github.com/konflux-ci/build-service/test/e2e/pkg/utils"
	"github.com/konflux-ci/build-service/test/e2e/pkg/utils/build"
)

var _ = framework.BuildSuiteDescribe("Build service E2E tests", Label("build-service"), func() {

	var f *framework.Framework
	AfterEach(framework.ReportFailure(&f))
	var err error
	defer GinkgoRecover()

	Describe("Creating component with container image source", Label("image-source", "github"), Ordered, func() {
		var applicationName, componentName, testNamespace string
		var timeout time.Duration
		var buildPipelineAnnotation map[string]string

		BeforeAll(func() {
			applicationName = fmt.Sprintf("test-app-%s", util.GenerateRandomString(4))
			f, err = framework.NewFramework(utils.GetGeneratedNamespace("build-e2e"))
			Expect(err).NotTo(HaveOccurred())
			testNamespace = f.UserNamespace

			_, err = f.AsKubeAdmin.HasController.CreateApplication(applicationName, testNamespace)
			Expect(err).NotTo(HaveOccurred())

			componentName = fmt.Sprintf("build-suite-test-component-image-source-%s", util.GenerateRandomString(6))
			timeout = time.Second * 10

			buildPipelineAnnotation = build.GetBuildPipelineBundleAnnotation(constants.DockerBuildOciTA)

			// Create a component with containerImageSource being defined.
			// A git source is required by the vcomponent.kb.io webhook,
			// but skip-initial-checks=true (passed below) prevents a build PipelineRun from triggering.
			component := appservice.ComponentSpec{
				ComponentName:  componentName,
				ContainerImage: containerImageSource,
				Source: appservice.ComponentSource{
					ComponentSourceUnion: appservice.ComponentSourceUnion{
						GitSource: &appservice.GitSource{
							URL: fmt.Sprintf(githubUrlFormat, githubOrg, pythonComponentRepoName),
						},
					},
				},
			}
			_, err = f.AsKubeAdmin.HasController.CreateComponent(component, testNamespace, containerImageSource, "", applicationName, true, buildPipelineAnnotation)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterAll(func() {
			if !CurrentSpecReport().Failed() {
				Eventually(func() error {
					return f.AsKubeAdmin.HasController.DeleteAllComponentsInASpecificNamespace(testNamespace, time.Minute*2)
				}, 2*time.Minute, 10*time.Second).Should(Succeed())
				Eventually(func() error {
					return f.AsKubeAdmin.HasController.DeleteAllApplicationsInASpecificNamespace(testNamespace, time.Minute*2)
				}, 2*time.Minute, 10*time.Second).Should(Succeed())
			}
		})

		It("should not trigger a PipelineRun", func() {
			Consistently(func() bool {
				_, err := f.AsKubeAdmin.HasController.GetComponentPipelineRun(componentName, applicationName, testNamespace, "")
				Expect(err).To(HaveOccurred())
				return strings.Contains(err.Error(), "no pipelinerun found")
			}, timeout, constants.PipelineRunPollingInterval).Should(BeTrue(), fmt.Sprintf("expected no PipelineRun to be triggered for the component %s in %s namespace", componentName, testNamespace))
		})
	})
})
