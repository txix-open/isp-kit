package soap

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
)

const (
	// ActionHeader is the HTTP header name for the SOAP action.
	ActionHeader = "SOAPAction"
)

// ActionMux routes SOAP requests based on the SOAPAction HTTP header.
// It provides a simple multiplexer for handling multiple SOAP operations.
type ActionMux struct {
	handlers map[string]http.Handler
}

// NewActionMux creates a new ActionMux with an empty handler map.
func NewActionMux() *ActionMux {
	return &ActionMux{handlers: map[string]http.Handler{}}
}

// Handle registers a handler for the specified SOAP action URI.
// It panics if a handler for the action is already registered.
// Returns the ActionMux for fluent chaining.
func (m *ActionMux) Handle(actionUri string, handler http.Handler) *ActionMux {
	_, ok := m.handlers[actionUri]
	if ok {
		panic(errors.Errorf("handler for action %v is already provided", actionUri))
	}
	m.handlers[actionUri] = handler
	return m
}

// ServeHTTP implements the http.Handler interface and routes requests based on SOAPAction.
// It returns a SOAP fault if the SOAPAction header is missing or the action is unknown.
// The action is added to the request context for logging purposes.
func (m *ActionMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	action := request.Header.Get(ActionHeader)
	if action == "" {
		_ = Fault{Code: "Client", String: "SOAPAction expected in http header"}.WriteError(writer)
		return
	}

	handler, ok := m.handlers[action]
	if !ok {
		_ = Fault{Code: "Client", String: fmt.Sprintf("unknown soap action: %s", action)}.WriteError(writer)
		return
	}

	ctx := log.ToContext(request.Context(), log.String("soapAction", action))
	request = request.WithContext(ctx)

	handler.ServeHTTP(writer, request)
}
