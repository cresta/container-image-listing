package containerimagelisting

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/pkg/errors"
)

// TODO Create ContainerClient factory method, IE NewContainerClient...etc
// Should be able to figure out correct client DockerHub/GHCR/ECR based on url
// TODO wrap errors from external librarys

const DockerHubMaxPageSize = 10 // TODO temporarily setting this to 10 for testing
const DockerHubBaseURL = "docker.io"
const GHCRBaseURL = "ghcr.io"
const ECRBaseURL = "amazonaws.com"

type DockerRegistryClient struct {
	Username string
	Password string
	BaseURL  string
}

var _ ContainerClient = &DockerRegistryClient{}

// parseBearerResponse - Parses bearer token from auth response
func (c *DockerRegistryClient) parseBearerResponse(body io.ReadCloser) (string, error) {
	type authResponse struct {
		Token       string    `json:"token"`
		AccessToken string    `json:"access_token"`
		ExpiresIn   int       `json:"expires_in"`
		IssuedAt    time.Time `json:"issued_at"`
	}

	ar := authResponse{}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(bodyBytes, &ar)
	if err != nil {
		return "", err
	}

	return ar.Token, nil
}

// isDockerHub - Helper function for quickly determining if something is docker.io
// docker.io uses different base URLs depending on the operation
// docker.io also requests specific headers in some instances
func (c *DockerRegistryClient) isDockerHub() bool {
	return strings.Contains(c.BaseURL, DockerHubBaseURL)
}

func (c *DockerRegistryClient) isECR() bool {
	return strings.Contains(c.BaseURL, ECRBaseURL)
}

func (c *DockerRegistryClient) getBearerECRSpecific() (token string, err error) {
	creds := credentials.NewEnvCredentials()
	config := aws.NewConfig().WithRegion("us-west-2").WithCredentials(creds)
	sess, err := session.NewSession(config)
	if err != nil {
		return "", err
	}
	svc := ecr.New(sess)

	input := &ecr.GetAuthorizationTokenInput{}

	result, err := svc.GetAuthorizationToken(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecr.ErrCodeServerException:
				fmt.Println(ecr.ErrCodeServerException, aerr.Error())
			case ecr.ErrCodeInvalidParameterException:
				fmt.Println(ecr.ErrCodeInvalidParameterException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return "", err
	}

	token = *result.AuthorizationData[0].AuthorizationToken
	return token, nil
}

// getBearerForRepo - Obtains a bearer token for a specific repository
// This is necessary because bearer tokens must be scoped for specific repositories
func (c *DockerRegistryClient) getBearerForRepo(name string) (token string, err error) {
	baseUrl := c.BaseURL
	if c.isDockerHub() {
		baseUrl = "auth.docker.io"
	}

	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://%s/token", baseUrl),
		nil)
	if err != nil {
		return "", err
	}

	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	q := req.URL.Query()

	if c.isDockerHub() {
		q.Add("service", "registry.docker.io")
	}
	q.Add("scope", fmt.Sprintf("repository:%s:pull", name))
	req.URL.RawQuery = q.Encode()

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	token, err = c.parseBearerResponse(resp.Body)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ListTags - Return tags for name in no particular order
// IE, name="library/redis"
func (c *DockerRegistryClient) ListTags(name string) ([]Tag, error) { // TODO have this return a struct of type Tag, use nil (or equivalent) for values we don't know yet
	return c.listTagsPage(name, 0)
}

func (c *DockerRegistryClient) listTagsPage(name string, page int) ([]Tag, error) {
	// Get auth token
	var token string
	var err error
	if c.isECR() {
		token, err = c.getBearerECRSpecific()
	} else {
		token, err = c.getBearerForRepo(name)
	}
	if err != nil {
		return nil, err
	}

	// Create URL
	baseUrl := c.BaseURL
	if strings.Contains(baseUrl, "docker.io") {
		baseUrl = "registry-1.docker.io"
	}
	url := fmt.Sprintf("https://%s/v2/%s/tags/list", baseUrl, name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add auth
	req.Header.Add("Accept", "application/json")
	authMethod := "Bearer"
	if c.isECR() {
		authMethod = "Basic"
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", authMethod, token))

	// Perform request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("status code was %d instead of 200 with status %s", resp.StatusCode, resp.Status)
		log.Print(msg)
		return nil, errors.New(msg)
	}

	// Parse tags
	type tagListResp struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	var tlr tagListResp

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bodyBytes, &tlr)
	if err != nil {
		return nil, err
	}

	// Convert tag string names to Tag structs
	tags := stringNamesToTags(tlr.Tags)

	return tags, nil
}