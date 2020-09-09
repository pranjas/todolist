package tptverify

import (
	"errors"
	"log"
	"todolist/handlers/token"
)

//Dummy Type that implements Verifier
type GoogleVerifier struct {
}

func (gverifier *GoogleVerifier) UserId(data interface{}) (string, error) {
	var googleClaims *token.GoogleClaim
	googleClaims, ok := data.(*token.GoogleClaim)
	if !ok {
		return "", errors.New("Not a GoogleClaim")
	}
	return googleClaims.Subject, nil
}

func (gverifier *GoogleVerifier) Verify(tokenString string) (interface{}, error) {
	googleClaims, err := token.GetGoogleClaims(tokenString)

	if err != nil {
		log.Printf("Invalid google token %s", tokenString)
		return nil, err
	}
	return googleClaims, nil
}

func (gverifier *GoogleVerifier) ResponseMap(data interface{}) map[string]interface{} {
	var googleClaims *token.GoogleClaim
	googleClaims, ok := data.(*token.GoogleClaim)
	if !ok {
		return nil
	}
	responseMap := map[string]interface{}{}
	responseMap["email"] = googleClaims.Email
	responseMap["email_verified"] = googleClaims.EmailVerified
	responseMap["picture"] = googleClaims.PictureURL
	responseMap["given_name"] = googleClaims.Firstname
	responseMap["family_name"] = googleClaims.Lastname
	responseMap["locale"] = googleClaims.Locale
	responseMap["userid"] = googleClaims.Subject
	return responseMap
}
