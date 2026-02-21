package providers

import (
	"context"
	"fmt"

	oidc "github.com/coreos/go-oidc"
)

type IDTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

type MultiVerifier []IDTokenVerifier

func (m MultiVerifier) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	if len(m) == 0 {
		return nil, fmt.Errorf("no ID token verifiers configured")
	}

	var lastErr error
	for _, verifier := range m {
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err == nil {
			return idToken, nil
		}
		lastErr = err
	}

	return nil, lastErr
}
