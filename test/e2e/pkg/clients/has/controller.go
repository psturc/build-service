package has

import (
	"fmt"

	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/forgejo"
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/github"
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/gitlab"
	"github.com/konflux-ci/build-service/test/e2e/pkg/clients/kube"
	"github.com/konflux-ci/build-service/test/e2e/pkg/constants"
	"github.com/konflux-ci/build-service/test/e2e/pkg/utils"
)

type Controller struct {
	GitHub  *github.Client
	GitLab  *gitlab.Client
	Forgejo *forgejo.ForgejoClient
	*kube.CustomClient
}

func NewController(k *kube.CustomClient) (*Controller, error) {
	gh, err := github.NewClient(utils.GetEnv(constants.GITHUB_TOKEN_ENV, ""),
		utils.GetEnv(constants.GITHUB_E2E_ORGANIZATION_ENV, "redhat-appstudio-qe"))
	if err != nil {
		return nil, err
	}

	groupId := utils.GetEnv("GITLAB_GROUP_ID", constants.DefaultGilabGroupId)
	gl, err := gitlab.NewClient(utils.GetEnv(constants.GITLAB_BOT_TOKEN_ENV, ""),
		utils.GetEnv(constants.GITLAB_API_URL_ENV, constants.DefaultGitLabAPIURL), groupId)
	if err != nil {
		return nil, err
	}

	var fj *forgejo.ForgejoClient
	forgejoToken := utils.GetEnv(constants.CODEBERG_BOT_TOKEN_ENV, "")
	if forgejoToken != "" {
		fj, err = forgejo.NewForgejoClient(
			forgejoToken,
			utils.GetEnv(constants.CODEBERG_API_URL_ENV, constants.DefaultCodebergAPIURL),
			utils.GetEnv(constants.CODEBERG_QE_ORG_ENV, constants.DefaultCodebergQEOrg),
		)
		if err != nil {
			fmt.Printf("WARNING: failed to authenticate with Forgejo/Codeberg in HasController (Forgejo retrigger will not work): %v\n", err)
			fj = nil
		}
	}

	return &Controller{
		GitHub:       gh,
		GitLab:       gl,
		Forgejo:      fj,
		CustomClient: k,
	}, nil
}
