package tptverify

import "github.com/pkg/errors"

type Verifier interface {
	//Verify would return something usable for
	//ResponseMap, usually it'll be the claims.
	Verify(token string) (interface{}, error)

	//ResponseMap
	//data is something which Verifier implementor expects.
	//Usually it'll be the claims.
	ResponseMap(data interface{}) map[string]interface{}

	UserId(data interface{}) (string, error)
}

func GetVerifier(authProvider string) (Verifier, error) {
	switch authProvider {
	case "google":
		return &GoogleVerifier{}, nil
	default:
		return nil, errors.Errorf("Provider %s not found", authProvider)
	}
}
