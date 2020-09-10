package test_helpers

import "net/http"

func StatusCodePoller(url string) func() int {
	return func() int {
		resp, err := http.Get(url)
		if err != nil {
			return 0
		}
		resp.Body.Close()
		return resp.StatusCode
	}
}
