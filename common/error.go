package common

import "fmt"

type IApiError interface {
	ErrorToJSON(int) string
	ErrorFromString(string) string
}

type ApiError struct{}

var errCodes = map[int]string{
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	500: "Service Error",
	503: "Service Unavailable",
	600: "Missing data source"}

func (err ApiError) ErrorToJSON(code int) string {
	return fmt.Sprintf("{\"%d\": \"%s\"}", code, errCodes[code])
}

func (err ApiError) ErrorFromString(message string) string {
	return fmt.Sprintf("{\"%d\": \"%s\"}", 500, message)
}
