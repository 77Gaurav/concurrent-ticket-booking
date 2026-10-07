package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrentBooking(t *testing.T) {
	const (
		totalRequests = 100
		eventID       = 1
		baseURL       = "http://localhost:8080"
	)

	var successCount int32
	var conflictCount int32
	var otherCount int32

	var wg sync.WaitGroup
	wg.Add(totalRequests)

	start := make(chan struct{})

	for i := 0; i < totalRequests; i++ {
		go func(i int) {
			defer wg.Done()

			// Make all requests start at roughly the same time.
			<-start

			payload := map[string]interface{}{
				"username": fmt.Sprintf("user-%d", i),
				"event_id": eventID,
			}

			body, err := json.Marshal(payload)
			if err != nil {
				t.Errorf("request %d: failed to marshal payload: %v", i, err)
				return
			}

			req, err := http.NewRequest(
				http.MethodPost,
				baseURL+"/bookings",
				bytes.NewReader(body),
			)
			if err != nil {
				t.Errorf("request %d: failed to create request: %v", i, err)
				return
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("request %d: failed to send request: %v", i, err)
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusCreated:
				atomic.AddInt32(&successCount, 1)

			case http.StatusConflict:
				atomic.AddInt32(&conflictCount, 1)

			default:
				atomic.AddInt32(&otherCount, 1)
				t.Errorf(
					"request %d: unexpected status code: %d",
					i,
					resp.StatusCode,
				)
			}
		}(i)
	}

	// Release all goroutines together.
	close(start)

	wg.Wait()

	t.Logf("total requests: %d", totalRequests)
	t.Logf("successful bookings: %d", successCount)
	t.Logf("conflicts: %d", conflictCount)
	t.Logf("other responses: %d", otherCount)

	if successCount != 5 {
		t.Fatalf(
			"expected exactly 5 successful bookings, got %d",
			successCount,
		)
	}

	if conflictCount != 95 {
		t.Fatalf(
			"expected exactly 95 rejected bookings, got %d",
			conflictCount,
		)
	}

	if otherCount != 0 {
		t.Fatalf(
			"expected 0 unexpected responses, got %d",
			otherCount,
		)
	}
}
