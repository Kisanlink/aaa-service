package model

type Response struct {
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	Error      any    `json:"error"`
	TimeStamp  string `json:"timestamp"`
}
