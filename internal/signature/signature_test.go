package signature_test

import (
	_ "embed"
	"github.com/kyma-project/api-gateway/internal/signature"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Correctly signed message should be signed off by key with public identity
// EDDSA 2A52DB2FC88744DB2B1742AD01F3FD39CA0C827E
//
//go:embed correctly_signed.sig
var correctlySignedSig []byte

// Message signed by impersonated key with public identity
// EDDSA B4877503B192609A2E22C81739FACBA528FDF429
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
