package notify

// API .
type API struct{}

// Request Object .
type Request struct {
	Message string `json:"message"`
}

// Response Object .
type Response struct {
	Message string `json:"message"`
}
