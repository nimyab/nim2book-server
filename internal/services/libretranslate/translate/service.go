package translate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/avast/retry-go/v4"
)

type HTTPClient interface {
	PostForm(url string, data url.Values) (*http.Response, error)
}

type Service struct {
	URL        string
	HTTPClient HTTPClient
}

type DefaultHTTPClient struct{}

func (c *DefaultHTTPClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return http.PostForm(url, data)
}

func New(libretranslateURL string, client HTTPClient) *Service {
	if client == nil {
		client = &DefaultHTTPClient{}
	}
	return &Service{
		URL:        libretranslateURL,
		HTTPClient: client,
	}
}

func (s *Service) Translate(input *Input) (*Output, error) {
	const operation = "translate.Service.Translate"

	formData := url.Values{}
	formData.Set("q", input.Q)
	formData.Set("source", string(input.Source))
	formData.Set("target", string(input.Target))
	if input.Format != nil {
		formData.Set("format", *input.Format)
	}
	if input.Alternatives != nil {
		formData.Set("alternatives", strconv.Itoa(*input.Alternatives))
	}

	resp, err := retry.DoWithData(func() (*http.Response, error) {
		return s.HTTPClient.PostForm(fmt.Sprintf("%s/%s", strings.TrimRight(s.URL, "/"), "translate"), formData)
	}, retry.Attempts(5))

	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("%s: status code: %d", operation, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	output := new(Output)
	if err = json.Unmarshal(respBody, output); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return output, nil
}
