package mixcloud

type Response struct {
	Details map[string][]string `json:"details,omitempty"`
	Error   *ErrorMessage       `json:"error,omitempty"`
	Result  *ResponseResult     `json:"result,omitempty"`
}

type ErrorMessage struct {
	Message    string `json:"message,omitempty"`
	Type       string `json:"type,omitempty"`
	RetryAfter int    `json:"retry_after,omitempty"`
}

type ResponseResult struct {
	Key     string `json:"key,omitempty"`
	Message string `json:"message,omitempty"`
	Success bool   `json:"success",omitempty`
}
