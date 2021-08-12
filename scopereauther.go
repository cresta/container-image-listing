package containerimagelisting

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type authRequest struct {
	Type   string
	Values map[string]string
}

var parserRegex = regexp.MustCompile(`,*([^"]*)="([^"]*)"`)

func parseWwwAuthenticate(requiredScope string) *authRequest {
	// Www-Authenticate is documented at https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate

	// WWW-Authenticate: <type> realm=<realm>[, charset="UTF-8"]

	// Expect a format like `Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:samalba/my-app:pull,push"`

	parts := strings.SplitN(requiredScope, " ", 2)
	if len(parts) != 2 {
		return nil
	}
	ret := authRequest{
		Type:   parts[0],
		Values: make(map[string]string),
	}
	scopes := parserRegex.FindAllStringSubmatch(parts[1], -1)
	for _, s := range scopes {
		if len(s) != 3 {
			panic("This is a logic error and should never happen!")
		}
		ret.Values[s[1]] = s[2]
	}
	return &ret
}

// ScopeReauther uses the dockerhubv2 auth API to execute another request for a token when an initial request fails
// due to a 4xx issues.
type ScopeReauther struct {
	Username string
	Password string
}

// Format documented on https://docs.docker.com/registry/spec/auth/token/
type authResponse struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

func (a *authResponse) tokenToUse() string {
	// Note: Spec says use any non empty token
	if a.Token != "" {
		return a.Token
	}
	return a.AccessToken
}

// RequestWrapper is any type that can wrap a request before it is executed
type RequestWrapper interface {
	Wrap(request *http.Request) error
}

type RequestWrapperFunc func(r *http.Request) error

func (r RequestWrapperFunc) Wrap(request *http.Request) error {
	return r(request)
}

var _ RequestWrapper = RequestWrapperFunc(nil)

// CheckForReauth returns a RequestWrapper for a response if the response is asking for authentication.  The returned
// RequestWrapper will usually set authorization headers
func (s *ScopeReauther) CheckForReauth(ctx context.Context, originalResp *http.Response, client *http.Client) (RequestWrapper, error) {
	// A need to auth should be inside the 4xx status code range
	if originalResp.StatusCode < 400 || originalResp.StatusCode > 499 {
		return nil, nil
	}
	authReq := originalResp.Header.Get("Www-Authenticate")
	if authReq == "" {
		return nil, nil
	}
	parsedRequest := parseWwwAuthenticate(authReq)
	if parsedRequest == nil {
		return nil, nil
	}
	newReqInto, err := url.Parse(parsedRequest.Values["realm"])
	if err != nil {
		return nil, fmt.Errorf("unable to parse realm URL: %w", err)
	}
	newQuery := make(url.Values)
	for k, v := range parsedRequest.Values {
		if k == "realm" || k == "" {
			continue
		}
		newQuery.Add(k, v)
	}
	newReqInto.RawQuery = newQuery.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, newReqInto.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build request: %w", err)
	}
	if s.Username != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to fetch auth context: %w", err)
	}
	var ret authResponse
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("uanble to decode response body as JSON: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("unable to close response body: %w", err)
	}
	return RequestWrapperFunc(func(req *http.Request) error {
		// TODO: Also check expiry token (???)
		req.Header.Set("Authorization", fmt.Sprintf("%s %s", parsedRequest.Type, ret.tokenToUse()))
		return nil
	}), nil
}
