package healthcheck

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/txix-open/isp-kit/json"
)

const (
	// cacheTime specifies how long to cache health check results before re-running them.
	cacheTime = 1 * time.Second
	// defaultHandleTimeout is the default timeout for the HTTP handler.
	defaultHandleTimeout = 1 * time.Second
)

// Registry manages a collection of health check components and provides an HTTP
// handler to expose their status. It caches results to avoid excessive checks.
//
// Registry is safe for concurrent use by multiple goroutines.
type Registry struct {
	checkers      map[string]Checker
	lastResult    Result
	lastCheckTime time.Time
	lock          sync.Locker
	handleTimeout time.Duration
}

// NewRegistry creates a new Registry with the specified handle timeout.
// If handleTimeout is zero, it defaults to 1 second.
func NewRegistry(handleTimeout time.Duration) *Registry {
	if handleTimeout == 0 {
		handleTimeout = defaultHandleTimeout
	}

	return &Registry{
		checkers:      make(map[string]Checker),
		lock:          &sync.Mutex{},
		handleTimeout: handleTimeout,
	}
}

// Register adds a health check component to the registry.
// The name parameter is used as the identifier for the component in the results.
// The checker parameter must implement the Checker interface.
func (r *Registry) Register(name string, checker Checker) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.checkers[name] = checker
}

// Unregister removes a health check component from the registry.
// If the component does not exist, this method has no effect.
func (r *Registry) Unregister(name string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.checkers, name)
}

// Handler returns an HTTP handler that exposes the health status of all
// registered components. The handler returns:
//   - 200 OK with status "pass" if all components are healthy
//   - 500 Internal Server Error with status "fail" if any component is unhealthy
//
// The response is encoded in application/health+json format according to
// the draft-inadarei-api-health-check specification.
func (r *Registry) Handler() http.Handler {
	return http.TimeoutHandler(http.HandlerFunc(r.handle), r.handleTimeout, "timeout")
}

// handle processes HTTP requests and returns the health check results.
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

// result computes and returns the current health check result.
// Results are cached for cacheTime to avoid running checks on every request.
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
