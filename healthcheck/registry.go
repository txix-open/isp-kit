package healthcheck

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/integration-system/isp-kit/json"
)

const (
	cacheTime = 1 * time.Second
)

type Registry struct {
	checkers      map[string]Checker
	lastResult    Result
	lastCheckTime time.Time
	lock          sync.Locker
}

func NewRegistry() *Registry {
	return &Registry{
		checkers: make(map[string]Checker),
		lock:     &sync.Mutex{},
	}
}

func (r *Registry) Register(name string, checker Checker) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.checkers[name] = checker
}

func (r *Registry) Handler() http.Handler {
	return http.TimeoutHandler(http.HandlerFunc(r.handle), 1*time.Second, "timeout")
}

func (r *Registry) handle(writer http.ResponseWriter, _ *http.Request) {
	result := r.result()

	statusCode := http.StatusOK
	if result.Status == StatusFail {
		statusCode = http.StatusInternalServerError
	}

	writer.Header().Set("Content-Type", "application/health+json")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(result)
}

// https://tools.ietf.org/id/draft-inadarei-api-health-check-01.html
func (r *Registry) result() Result {
	r.lock.Lock()
	defer r.lock.Unlock()

	now := time.Now().UTC()
	if now.Before(r.lastCheckTime.Add(cacheTime)) {
		return r.lastResult
	}

	details := make(map[string][]Detail)
	resultStatus := StatusPass
	for name, checker := range r.checkers {
		resultErr := checker.Healthcheck(context.Background())
		status := StatusPass
		output := ""
		if resultErr != nil {
			resultStatus = StatusFail
			status = StatusFail
			output = resultErr.Error()
		}
		detail := Detail{
			Status:        status,
			Output:        output,
			ComponentName: name,
			ComponentType: "component",
			Time:          time.Now().UTC(),
		}
		details[name] = []Detail{detail}
	}
	result := Result{
		Status:  resultStatus,
		Details: details,
	}

	r.lastCheckTime = now
	r.lastResult = result

	return result
}
