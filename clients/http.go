package clients

import (
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Duration(5) * time.Second,
}

// GetHTTPClient returns a single HttpClient for use across the app
func GetHTTPClient() (*http.Client)  {
	return httpClient
}