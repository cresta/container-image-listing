package containerimagelisting

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ecr"
)

// ECRClient should connect to AWS's ECR for Docker tokens
type ECRClient interface {
	// GetAuthorizationTokenWithContext should emulate aws-sdk-go's ECR.GetAuthorizationTokenWithContext function
	GetAuthorizationTokenWithContext(ctx aws.Context, input *ecr.GetAuthorizationTokenInput, opts ...request.Option) (*ecr.GetAuthorizationTokenOutput, error)
}

var _ ECRClient = &ecr.ECR{}

// ECRAuthWrapper can wrap http.Request with the ECR Docker authentication token
type ECRAuthWrapper struct {
	ECR                     ECRClient
	AuthBufferTime          time.Duration
	cachedAuthorizationData *ecr.AuthorizationData
	mu                      sync.Mutex
}

func (a *ECRAuthWrapper) authBufferTime() time.Duration {
	if a.AuthBufferTime == 0 {
		return time.Minute
	}
	return a.AuthBufferTime
}

var _ RequestWrapper = &ECRAuthWrapper{}

// Wrap a http.Request with the docker token.  If the token is unknown or expired, will fetch it before wrapping.
func (a *ECRAuthWrapper) Wrap(request *http.Request) error {
	token, err := a.FetchToken(request.Context())
	if err != nil {
		return fmt.Errorf("unable to fetch request token for ECR: %w", err)
	}
	request.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))
	return nil
}

// FetchToken returns the ECR docker token.  It's possible to call this before using ECRAuthWrapper to verify
// you are able to fetch a token.
func (a *ECRAuthWrapper) FetchToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cachedAuthorizationData != nil {
		if a.cachedAuthorizationData.ExpiresAt.After(time.Now().Add(a.authBufferTime())) {
			return *a.cachedAuthorizationData.AuthorizationToken, nil
		}
	}
	var input ecr.GetAuthorizationTokenInput

	result, err := a.ECR.GetAuthorizationTokenWithContext(ctx, &input)
	if err != nil {
		return "", fmt.Errorf("error getting ECR authorization token: %w", err)
	}
	if len(result.AuthorizationData) < 1 {
		return "", fmt.Errorf("unexpected return from ECR, expected at least one token, but got zero")
	}
	a.cachedAuthorizationData = result.AuthorizationData[0]

	return *a.cachedAuthorizationData.AuthorizationToken, nil
}
