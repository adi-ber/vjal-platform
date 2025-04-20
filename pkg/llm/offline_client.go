package llm

import "context"

// OfflineClient is a stub for an embedded/offline model.
type OfflineClient struct{}

// NewOfflineClient constructs an OfflineClient.
func NewOfflineClient(cfg map[string]string) (Client, error) {
	return &OfflineClient{}, nil
}

// Prompt returns a simple stub response.
func (o *OfflineClient) Prompt(ctx context.Context, prompt string) (string, error) {
	return "Offline response to: " + prompt, nil
}

// HealthCheck for OfflineClient is always healthy.
func (o *OfflineClient) HealthCheck(ctx context.Context) error {
	return nil
}
