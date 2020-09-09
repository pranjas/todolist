package handlers

import "fmt"

const (
	API_ERROR_CODE_OK            int64 = 0
	API_ERROR_CODE_GENERIC_ERROR int64 = 0xebadf + iota
	API_ERROR_CODE_TOKEN_EXPIRED
	API_ERROR_CODE_INVALID_INPUT
	API_ERROR_CODE_INVALID_VERSION
	API_ERROR_CODE_UNKNOWN_AUTH_PROVIDER
	API_ERROR_CODE_INVALID_PUBLIC_CERT
	API_ERROR_CODE_INVALID_RSA_KEY
	API_ERROR_CODE_NOT_IMPLEMENTED
)

func ApiErrorCodeToString(errorCode int64) string {
	switch errorCode {
	case API_ERROR_CODE_GENERIC_ERROR:
		return "request contained errors. Check response headers"
	case API_ERROR_CODE_TOKEN_EXPIRED:
		return "token expired"
	case API_ERROR_CODE_INVALID_INPUT:
		return "invalid input in request"
	case API_ERROR_CODE_UNKNOWN_AUTH_PROVIDER:
		return "unknown auth provider"
	case API_ERROR_CODE_INVALID_PUBLIC_CERT:
		return "invalid public cert format"
	case API_ERROR_CODE_INVALID_RSA_KEY:
		return "invalid RSA key format"
	case API_ERROR_CODE_NOT_IMPLEMENTED:
		return "api is not currently implemented"
	default:
		return fmt.Sprintf("unknown error code %d", errorCode)
	}
}
