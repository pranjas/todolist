package tptverify

import "github.com/pkg/errors"

const (
	VERIFIER_GOOGLE   = "google"
	VERIFIER_FACEBOOK = "facebook"
	VERIFIER_US       = "us"
)

type Verifier interface {
	//Verify would return something usable for
	//ResponseMap, usually it'll be the claims.
	Verify(token string) (interface{}, error)

	//ResponseMap
	//data is something which Verifier implementor expects.
	//Usually it'll be the claims.
	ResponseMap(data interface{}) map[string]interface{}

	UserId(data interface{}) (string, error)
	Name() string
}

func GetVerifier(authProvider string) (Verifier, error) {
	switch authProvider {
	case VERIFIER_GOOGLE:
		return &GoogleVerifier{}, nil
	case VERIFIER_US:
		return &LocalVerifier{}, nil
	default:
		return nil, errors.Errorf("Provider %s not found", authProvider)
	}
}
