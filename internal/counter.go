package internal

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/onsi/ginkgo/types"
)

type Counter struct {
	Index int `json:"index"`
}

func MakeNextIndexCounter(suiteConfig types.SuiteConfig) func() (int, error) {
	if suiteConfig.ParallelTotal > 1 {
		client := &http.Client{}
		return func() (int, error) {
			resp, err := client.Get(suiteConfig.ParallelHost + "/counter")
			if err != nil {
				return -1, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return -1, fmt.Errorf("unexpected status code %d", resp.StatusCode)
			}

			var counter Counter
			err = json.NewDecoder(resp.Body).Decode(&counter)
			if err != nil {
				return -1, err
			}

			return counter.Index, nil
		}
	} else {
		idx := -1
		return func() (int, error) {
			idx += 1
			return idx, nil
		}
	}
}
