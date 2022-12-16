package responses

type Response struct {
	Status             int                    `json:"status,omitempty"`
	APICode            int64                  `json:"apicode,omitempty"`
	APICodeDescription string                 `json:"apicode_desc,omitempty"`
	Message            string                 `json:"msg,omitempty"`
	Meta               map[string]interface{} `json:"extra,omitempty"`
}
