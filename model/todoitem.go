package model

type TodoItem struct {
	Owner     string                 `json:"-"`
	Name      string                 `json:"name"`
	Content   map[string]interface{} `json:"content,omitempty"`
	Actions   map[string]interface{} `json:"actions",omitempty`
	StartTime string                 `json:"start_time",omitempty`
	EndTime   string                 `json:"end_time", omitempty`
}
