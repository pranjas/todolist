package responses

type Response struct {
	Status  int                    `json:"status",omitempty`
	Message string                 `json:"msg",omitempty`
	Meta    map[string]interface{} `json:"extra",omitempty`
}
