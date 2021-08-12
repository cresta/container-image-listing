package containerimagelisting

import (
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/stretchr/testify/require"
)

type TestingECRClient struct{}

func (t *TestingECRClient) GetAuthorizationTokenWithContext(_ aws.Context, _ *ecr.GetAuthorizationTokenInput, _ ...request.Option) (*ecr.GetAuthorizationTokenOutput, error) {
	return &ecr.GetAuthorizationTokenOutput{
		AuthorizationData: []*ecr.AuthorizationData{
			{
				AuthorizationToken: aws.String("test_token"),
				ExpiresAt:          aws.Time(time.Now().Add(time.Minute)),
			},
		},
	}, nil
}

func TestECRAuthWrapper_Wrap(t *testing.T) {
	a := ECRAuthWrapper{
		ECR: &TestingECRClient{},
	}
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	require.NoError(t, a.Wrap(req))
	require.Equal(t, "Basic test_token", req.Header.Get("Authorization"))
}
