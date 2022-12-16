package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"todolist/handlers/token"
	"todolist/responses"
	"todolist/tptverify"
	"todolist/utils"
)

//Map the required headers based on whether they are mandatory or not.
//Most of the time a header would contain only a single value but we
//make it an array here to allow for multiple values if any.

type requestHeaderChecker func(requestValues []string, possibleValues []string) bool
type requestHeadersRequired struct {
	Name      string
	Values    []string
	Required  bool
	CheckerFn requestHeaderChecker
}

type RequestClaims struct {
	Claims interface{}
	tptverify.Verifier
}

//Authorization checker is a bit special we need to check
//if Bearer is present.
func authorizationCheckerFunc(requestValues []string, possibleValues []string) bool {
	reqValue := strings.Split(requestValues[0], " ")[0]
	return reqValue == possibleValues[0]
}

func contentTypeCheckerFunc(requestValues []string, possibleValues []string) bool {
	//Check for application/json in Content-Type
	//whose value is as Content-Type: application/json; charset=UT-8
	reqValue := strings.Split(requestValues[0], ";")
	for v := range reqValue {
		for validValue := range possibleValues {
			if v == validValue {
				return true
			}
		}
	}
	return false
}

var headersRequired = map[string][]requestHeadersRequired{
	http.MethodGet: {
		{Name: "Content-Type", Values: []string{"application/json"}, Required: true, CheckerFn: contentTypeCheckerFunc},
	},
	http.MethodPost: {
		{Name: "Content-Type", Values: []string{"application/json"}, Required: true, CheckerFn: contentTypeCheckerFunc},
		{Name: "Authorization", Values: []string{"Bearer"}, Required: true, CheckerFn: authorizationCheckerFunc},
		{Name: "X-Resource-Auth", Values: []string{""}, Required: false},
	},
}

func GenericNotImplemented(w http.ResponseWriter, r *http.Request) {
	GenericResponseWithEC(&w, "not implemented",
		http.StatusNotImplemented, API_ERROR_CODE_GENERIC_ERROR)
}

func GenericResponse(w *http.ResponseWriter, message string, httpStatusCode int) {
	apiErrorCode := API_ERROR_CODE_OK
	if httpStatusCode != http.StatusOK {
		apiErrorCode = API_ERROR_CODE_GENERIC_ERROR
	}
	GenericResponseWithEC(w, message, httpStatusCode, apiErrorCode)
}

func GenericResponseWithEC(w *http.ResponseWriter, message string, httpStatusCode int, errorCode int64) {
	response := responses.Response{
		Status:             httpStatusCode,
		Message:            message,
		APICode:            errorCode,
		APICodeDescription: ApiErrorCodeToString(errorCode),
	}
	GenericWriteResponse(w, &response)
}

func GenericInternalServerError(w *http.ResponseWriter, message string) {
	GenericResponse(w, message, http.StatusInternalServerError)
}

func GenericBadRequest(w *http.ResponseWriter, message string) {
	GenericResponse(w, message, http.StatusBadRequest)
}

func GenericWriteResponse(w *http.ResponseWriter, resp *responses.Response) {
	if resp.APICode == API_ERROR_CODE_OK {
		if resp.Status != http.StatusOK {
			resp.APICode = API_ERROR_CODE_GENERIC_ERROR
		} else {
			resp.APICode = API_ERROR_CODE_OK
		}
	}
	if resp.APICodeDescription == "" {
		resp.APICodeDescription = ApiErrorCodeToString(resp.APICode)
	}
	respBytes, err := json.Marshal(*resp)
	if err != nil {
		(*w).WriteHeader(http.StatusInternalServerError)
		log.Printf("error Marshalling response %v", resp)
		return
	}
	(*w).Header().Set("Content-Type", "application/json; charset=UTF-8")
	(*w).WriteHeader(resp.Status)
	(*w).Write(respBytes)
}

func GenericInternalServerHeader(w *http.ResponseWriter, r *http.Request) {
	GenericWriteHeader(w, r, http.StatusInternalServerError)
}

func GenericWriteHeader(w *http.ResponseWriter, r *http.Request, code int) {
	(*w).WriteHeader(code)
}

//Heroku gives a header named X-Forwarded-Proto
//which contains the scheme the request originally
//landed on heroku server. NOTE that we don't run
//a HTTPS server, all requests come to us as plain
//HTTP request since it's forwarded internally by
//Heroku to us.
func redirectToHTTPS(w *http.ResponseWriter, r *http.Request) bool {
	scheme, ok := r.Header[utils.HerokuForwardedProto]
	//We're not running behind Heroku or a
	//Cloud based host that supports X-Forwarded-Proto.
	if !ok {
		return false
	}
	//The magic http code is 307 which causes http clients
	//to re-issue request with the correct http method.
	//Not using 307 causes some clients to change the original
	//http method to POST by default.
	if scheme[0] != "https" {
		httpsURL := fmt.Sprintf("https://%s%s", r.Host, r.URL.Path)
		if len(r.URL.RawQuery) > 0 {
			httpsURL = fmt.Sprintf("%s?%s", httpsURL, r.URL.RawQuery)
		}
		http.Redirect(*w, r, httpsURL, http.StatusTemporaryRedirect)
		return true
	}
	return false
}

func checkRequestHeaders(w *http.ResponseWriter, r *http.Request) bool {
	expectedHeaders, ok := headersRequired[r.Method]
	result := true
	if !ok {
		result = false
		goto out
	}
	//For a particular request, go over all expected headers.
	//and match the values.
	for _, header := range expectedHeaders {
		requestHeaderValues, ok := r.Header[header.Name]
		//Request doesn't contains the header.
		if !ok {
			//header not found but is required.
			if header.Required {
				result = false
				goto out
			}
			continue
		}
		//Support only the first value
		//check of request header
		if header.CheckerFn != nil {
			if header.Required && !header.CheckerFn(requestHeaderValues,
				header.Values) {
				result = false
				log.Printf("Header %s, found = %s, need = %s\n",
					header.Name, requestHeaderValues[0],
					header.Values[0])
				goto out
			}
		} else if (requestHeaderValues[0] != header.Values[0]) && header.Required {
			log.Printf("Header %s, found = %s, need = %s\n",
				header.Name, requestHeaderValues[0],
				header.Values[0])
			result = false
			goto out
		}
	}
out:
	if !result {
		response := responses.Response{
			Status:  http.StatusBadRequest,
			Message: "Required Request headers missing.",
		}
		GenericWriteResponse(w, &response)
	}
	return result
}

func checkRequestMethod(w *http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		response := responses.Response{
			Status:  http.StatusBadRequest,
			Message: "HTTP Method Not Supported",
		}
		GenericWriteResponse(w, &response)
		return false
	}
	return true
}

func VerifyBearerToken(request *http.Request) (*RequestClaims, responses.Response) {
	var result = &RequestClaims{}
	response := responses.Response{
		Status:             http.StatusOK,
		Message:            "",
		APICode:            API_ERROR_CODE_OK,
		APICodeDescription: ApiErrorCodeToString(API_ERROR_CODE_OK),
	}
	bearerToken := token.GetBearerToken(request)
	authProvider, err := utils.GetRequestHeader(request, "X-Resource-Auth")
	if err != nil {
		log.Printf("Verifier is not third-party\n")
		authProvider = tptverify.VERIFIER_US
	}
	provider, err := tptverify.GetVerifier(authProvider)
	if err != nil {
		response.Status = http.StatusBadRequest
		response.APICode = API_ERROR_CODE_INVALID_INPUT
		response.APICodeDescription = ApiErrorCodeToString(response.APICode)
		log.Printf("Error = %s", err)
		return result, response
	}
	result.Verifier = provider

	claims, err := provider.Verify(bearerToken)
	if err != nil {
		response.Status = http.StatusBadRequest
		response.APICode = API_ERROR_CODE_TOKEN_EXPIRED
		response.APICodeDescription = ApiErrorCodeToString(response.APICode)
		log.Printf("Error = %s", err)
	} else {
		result.Claims = claims
	}
	return result, response
}
