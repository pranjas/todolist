package token

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"todolist/environment"
	"todolist/model"

	"github.com/dgrijalva/jwt-go"
)

const (
	maxExpirationMins = time.Minute * 15
)

//We'll store the userid currently.
//AppClaim is a wrapper over the standard
//jwt claim.
type AppClaim struct {
	Resource  string          `json:"res"`
	Id        string          `json:"id"`
	TokenType model.LoginType `json:"token_type"`
	jwt.StandardClaims
}

func getSigningKey() string {
	return base64.StdEncoding.EncodeToString([]byte(environment.GetAppTokenSecret()))
}

func GenerateTokenWithTimeout(user *model.User, timeout int64, tokenType model.LoginType) (string, error) {
	appClaim := AppClaim{
		Resource: "login",
		Id:       user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: timeout,
			Issuer:    environment.GetAppName(),
			NotBefore: time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, appClaim)
	signedToken, err := token.SignedString([]byte(getSigningKey()))
	if err != nil {
		return "", err
	}
	log.Printf("Generated %s for user %s", signedToken, user.ID)
	return signedToken, nil
}

func GetBearerToken(r *http.Request) string {
	var bearerToken = ""
	if r.Header != nil {
		headerValues := r.Header["Authorization"]
		if len(headerValues) > 0 {
			authValues := strings.Split(headerValues[0], " ")
			if len(authValues) > 1 {
				bearerToken = authValues[1]
			}
		}
	}
	return bearerToken
}

func GenerateToken(user *model.User, loginType model.LoginType) (string, error) {
	return GenerateTokenWithTimeout(user, time.Now().Add(maxExpirationMins).Unix(), loginType)
}

func GetUserClaims(tokenString string) (*AppClaim, error) {
	appClaim := AppClaim{}

	token, err := jwt.ParseWithClaims(tokenString, &appClaim, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return ([]byte(getSigningKey())), nil
	})
	if err != nil {
		log.Printf("Error validating token")
		return nil, err
	}
	if claims, ok := token.Claims.(*AppClaim); ok && token.Valid {
		log.Printf("Claims are valid %v", claims)
		return &appClaim, nil
	}
	return nil, nil
}
