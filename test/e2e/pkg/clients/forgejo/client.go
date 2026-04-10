package forgejo

import (
	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
)

type ForgejoClient struct {
	client *forgejo.Client
	org    string
	apiURL string
	token  string
}

func NewForgejoClient(accessToken, baseURL, org string) (*ForgejoClient, error) {
	client, err := forgejo.NewClient(baseURL, forgejo.SetToken(accessToken))
	if err != nil {
		return nil, err
	}

	return &ForgejoClient{
		client: client,
		org:    org,
		apiURL: baseURL,
		token:  accessToken,
	}, nil
}

func (fc *ForgejoClient) GetClient() *forgejo.Client {
	return fc.client
}

func (fc *ForgejoClient) GetOrg() string {
	return fc.org
}
