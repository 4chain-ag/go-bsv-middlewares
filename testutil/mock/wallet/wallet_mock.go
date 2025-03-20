package mock

import (
	"context"
	"errors"

	"github.com/4chain-ag/go-bsv-middlewares/pkg/wallet"
)

// MockWallet provides a simple mock implementation of WalletInterface.
type MockWallet struct {
	identityKey string
	keyDeriver  bool
	validNonces map[string]bool
}

// NewMockWallet creates a new mock wallet with or without keyDeriver.
func NewMockWallet(enableKeyDeriver bool) wallet.WalletInterface {
	return &MockWallet{
		identityKey: IdentityKeyMock,
		keyDeriver:  enableKeyDeriver,
		validNonces: make(map[string]bool),
	}
}

// GetPublicKey returns a mock public key while validating required parameters.
func (m *MockWallet) GetPublicKey(ctx context.Context, options wallet.GetPublicKeyOptions) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	if options.Privileged {
		return "", errors.New(ErrorNoPrivilege)
	}

	if options.IdentityKey {
		if !m.keyDeriver {
			return "", errors.New(ErrorKeyDeriver)
		}
		return m.identityKey, nil
	}

	if options.ProtocolID == nil || options.KeyID == "" || options.KeyID == " " {
		return "", errors.New(ErrorMissingParams)
	}

	if !m.keyDeriver {
		return "", errors.New(ErrorKeyDeriver)
	}

	return DerivedKeyMock, nil
}

// CreateSignature returns a mock signature.
func (m *MockWallet) CreateSignature(ctx context.Context, data []byte, protocolID interface{}, keyID string, counterparty string) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if len(data) == 0 || keyID == "" || counterparty == "" {
		return nil, errors.New(ErrorInvalidInput)
	}

	return []byte(MockSignature), nil
}

// VerifySignature returns true if the signature matches expected mock data.
func (m *MockWallet) VerifySignature(ctx context.Context, data []byte, signature []byte, protocolID interface{}, keyID string, counterparty string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	return string(signature) == MockSignature, nil
}

// CreateNonce generates a deterministic nonce.
func (m *MockWallet) CreateNonce(ctx context.Context) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	m.validNonces[MockNonce] = true
	return MockNonce, nil
}

// VerifyNonce checks if the nonce exists.
func (m *MockWallet) VerifyNonce(ctx context.Context, nonce string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	_, exists := m.validNonces[nonce]
	return exists, nil
}

// ListCertificates returns an empty list.
func (m *MockWallet) ListCertificates(ctx context.Context, certifiers []string, types []string) ([]wallet.Certificate, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return []wallet.Certificate{}, nil
}

// ProveCertificate returns an empty map.
func (m *MockWallet) ProveCertificate(ctx context.Context, certificate wallet.Certificate, verifier string, fieldsToReveal []string) (map[string]string, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return map[string]string{}, nil
}
