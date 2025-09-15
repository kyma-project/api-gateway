package signature

import (
	_ "embed"
	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

//go:embed pub_key.pgp
var publicKey string

func DecryptAndVerifySignature(data []byte) (string, bool, error) {
	pgp := crypto.PGP()
	keyObj, err := crypto.NewKeyFromArmored(publicKey)
	if err != nil {
		return "", false, err
	}

	verifier, err := pgp.Verify().
		VerificationKey(keyObj).
		New()
	if err != nil {
		return "", false, err
	}
	verifyResult, err := verifier.VerifyInline(data, crypto.Bytes)
	if err != nil {
		return "", false, err
	}

	if sigErr := verifyResult.SignatureError(); sigErr != nil {
		return "", false, sigErr
	}

	return verifyResult.String(), true, nil
}
