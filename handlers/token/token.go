package token

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"todolist/environment"
	"todolist/model"

	"github.com/dgrijalva/jwt-go"
)

const (
	maxExpirationMins    = time.Minute * 15
	googleCertificateURL = "https://www.googleapis.com/oauth2/v1/certs"
)

type googleTokenVerifier struct {
	certs   map[string]interface{}
	timeout int64
	sync.Mutex
}

var googleToken googleTokenVerifier

func init() {
	googleToken = googleTokenVerifier{}
}

//We'll store the userid currently.
//AppClaim is a wrapper over the standard
//jwt claim.
type AppClaim struct {
	Resource  string          `json:"res"`
	Id        string          `json:"id"`
	TokenType model.LoginType `json:"token_type"`
	jwt.StandardClaims
}

type GoogleClaim struct {
	jwt.StandardClaims
	//The following seven fields are available if the user has
	//granted the profile and email OAuth Scopes
	Email         string `json:"email",omitempty`
	EmailVerified bool   `json:"email_verified",omitempty`
	Name          string `json:"name",omitempty`
	PictureURL    string `json:"picture",omitempty`
	Firstname     string `json:"given_name",omitempty`
	Lastname      string `json:"family_name",omitempty`
	Locale        string `json:"locale",omitempty`
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
			Subject:   user.ID,
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

func GetGoogleClaims(tokenString string) (*GoogleClaim, error) {
	googleClaim := GoogleClaim{}
	googleToken.Lock()
	defer googleToken.Unlock()
check_again:
	if googleToken.certs == nil || len(googleToken.certs) == 0 {
		resp, err := http.Get(googleCertificateURL)
		if err != nil {
			log.Printf("Error connecting to %s", googleCertificateURL)
			return nil, err
		}
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading from %s", googleCertificateURL)
			return nil, err
		}
		err = json.Unmarshal(bytes, &googleToken.certs)
		if err != nil {
			log.Printf("Unable to unmarshal from %s", googleCertificateURL)
			return nil, err
		}
		for _, value := range resp.Header.Values("Cache-Control") {
			if strings.Contains(value, "max-age") {
				allValues := strings.Split(value, ",")
				var maxAgeVal string
				for _, val := range allValues {
					if strings.Contains(val, "max-age") {
						maxAgeVal = val
						break
					}
				}
				val, err := strconv.ParseInt(strings.Split(maxAgeVal, "=")[1], 10, 64)
				if err != nil {
					val = 0
				}
				log.Printf("Setting google's cert timeout value to %d seconds\n", val)
				googleToken.timeout = time.Now().Unix() + val
			}
		}
	} else if googleToken.timeout <= time.Now().Unix() {
		googleToken.certs = nil
		goto check_again
	}
	//ParseWithClaims requires either a signing key or
	//the public key. Since we're using the PEM format
	//the public key needs to be extraced from the certificate
	//presented by google.
	token, err := jwt.ParseWithClaims(tokenString, &googleClaim,
		func(token *jwt.Token) (interface{}, error) {
			var publicCert string
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
			}
			//Get the public key from the certificate
			//which signed this token. Which certificate
			//to use is identified by the "kid" field in
			//token. (kid = key identifier)
			publicCert = googleToken.certs[token.Header["kid"].(string)].(string)
			publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicCert))
			if err != nil {
				log.Printf("Not a valid RSA Key %s\n", publicCert)
				return nil, errors.New("Invalid RSA key found")
			}
			return publicKey, nil
		})
	if err != nil {
		log.Printf("error parsing %v\n", err)
		log.Printf("token = %v\n", token)
		return nil, errors.New("Invalid google token")
	}
	if claims, ok := token.Claims.(*GoogleClaim); ok && token.Valid {
		log.Printf("Claims are valid %v", claims)
		return &googleClaim, nil
	}
	return nil, errors.New("Invalid google token")
}
