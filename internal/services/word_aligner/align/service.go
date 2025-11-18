package align

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/avast/retry-go/v4"
)

type Service struct {
	URL string
}

var service *Service

// New creates a new instance of the Word Aligner service client.
// Deprecated: use Grpc client instead
func New(wordAlignerURL string) *Service {
	service = &Service{
		URL: wordAlignerURL,
	}
	return service
}

func (s *Service) Align(input *Input) (*Output, error) {
	const operation = "align.Service.Align"

	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	resp, err := retry.DoWithData(func() (*http.Response, error) {
		return http.Post(fmt.Sprintf("%s/%s", s.URL, "align"), "application/json", bytes.NewBuffer(data))
	}, retry.Attempts(5))

	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(resp.Body)
		slog.Debug("Align failed", "Status Code", resp.StatusCode, "Response", string(responseBody))
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
