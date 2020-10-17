package tptverify

import (
	"log"
	"todolist/handlers/token"

	"github.com/pkg/errors"
)

type LocalVerifier struct {
}

func (localVerifier *LocalVerifier) Name() string {
	return "self-verifier"
}

func (localVerifier *LocalVerifier) UserId(data interface{}) (string, error) {
	var appClaim *token.AppClaim
	appClaim, ok := data.(*token.AppClaim)
	if !ok {
		return "", errors.Errorf("Not a local claim.")
	}
	return appClaim.Id, nil
}

func (localVerifier *LocalVerifier) Verify(tokenString string) (interface{}, error) {
	appClaims, err := token.GetUserClaims(tokenString)
	if err != nil {
		log.Printf("Invalid App Claim token %s\n", tokenString)
		return nil, err
	}
	return appClaims, nil
}

func (localVerifier *LocalVerifier) ResponseMap(data interface{}) map[string]interface{} {
	//Return an empty map for now.
	//We don't have anything for now to send back.
	responseMap := map[string]interface{}{}
	return responseMap
}
