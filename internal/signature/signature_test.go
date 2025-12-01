package signature_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/api-gateway/internal/signature"
)

// Correctly signed message should be signed off by key with public hex identity
// fbcbd0163b221dca
//
//go:embed correctly_signed.sig
var correctlySignedSig []byte

// Message signed by impersonated key with public hex identity
// ddff4b9544cdb6c7
//
//go:embed signed_by_impersonated_key.sig
var signedByImpersonatedKeySig []byte

func TestDecryptAndVerifySignature(t *testing.T) {
	t.Run("should succeed for correctly signed message", func(t *testing.T) {
		t.Parallel()

		msg, valid, err := signature.DecryptAndVerifySignature(correctlySignedSig)
		assert.NoError(t, err)
		assert.True(t, valid)
		assert.Contains(t, msg, "test-shoot")
	})

	t.Run("should fail for message signed by impersonated key", func(t *testing.T) {
		t.Parallel()

		_, valid, err := signature.DecryptAndVerifySignature(signedByImpersonatedKeySig)
		assert.Error(t, err)
		assert.False(t, valid)
	})
}
