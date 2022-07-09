package pull_request

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"os"

	"code.gitea.io/sdk/gitea"
)

type GiteaService struct {
	client *gitea.Client
	owner  string
	repo   string
}

var _ PullRequestService = (*GiteaService)(nil)

func NewGiteaService(ctx context.Context, token, url, owner, repo string, insecure bool) (PullRequestService, error) {
	if token == "" {
		token = os.Getenv("GITEA_TOKEN")
	}
	httpClient := &http.Client{}
	if insecure {
		cookieJar, _ := cookiejar.New(nil)

		httpClient = &http.Client{
			Jar: cookieJar,
			/* False positive: All network communication is performed over TLS including service-to-service communication between the three components (argocd-server, argocd-repo-server, argocd-application-controller).

			Argo CD provides two inbound TLS endpoints that can be configured: (1) the user-facing endpoint of the argocd-server workload which serves the UI and the API and (2) the endpoint of the argocd-repo-server, 
			which is accessed by argocd-server and argocd-application-controller workloads to request repository operations. By default, and without further configuration, both of these endpoints will be set-up to use 
			an automatically generated, self-signed certificate. */
			
			/* #nosec G402 */
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}}
	}
	client, err := gitea.NewClient(url, gitea.SetToken(token), gitea.SetHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &GiteaService{
		client: client,
		owner:  owner,
		repo:   repo,
	}, nil
}

func (g *GiteaService) List(ctx context.Context) ([]*PullRequest, error) {
	opts := gitea.ListPullRequestsOptions{
		State: gitea.StateOpen,
	}
	prs, _, err := g.client.ListRepoPullRequests(g.owner, g.repo, opts)
	if err != nil {
		return nil, err
	}
	list := []*PullRequest{}
	for _, pr := range prs {
		list = append(list, &PullRequest{
			Number:  int(pr.Index),
			Branch:  pr.Head.Ref,
			HeadSHA: pr.Head.Sha,
		})
	}
	return list, nil
}
