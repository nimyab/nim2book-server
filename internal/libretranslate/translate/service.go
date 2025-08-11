package translate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/avast/retry-go/v4"
)

type Service struct {
	ULR string
}

var service *Service

func New(libretranslateURL string) *Service {
	service = &Service{
		ULR: libretranslateURL,
	}
	return service
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
		return http.PostForm(fmt.Sprintf("%s/%s", s.ULR, "translate"), formData)
	}, retry.Attempts(5))

	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("%s: status code: %d", operation, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer resp.Body.Close()

	output := new(Output)
	if err = json.Unmarshal(respBody, output); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return output, nil
}
