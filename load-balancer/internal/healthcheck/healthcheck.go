package healthcheck

import (
	"log"
	"net/http"
	"time"
)

// HTTP GET check
func CheckHTTP(url string, timeout time.Duration) bool {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Health check for %s failed: %v", url, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Health check for %s returned HTTP status code: %d", url, resp.StatusCode)
		return false
	}
	return true
}