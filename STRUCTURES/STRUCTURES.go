package STRUCTURES

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Status string `json:"status"`
}

type InternalResult struct {
	Json     string
	Is_error int
	Is_empty int
}
