package web

type WebResponse struct {
	Status  string      `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
