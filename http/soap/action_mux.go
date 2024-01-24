package soap

import (
	"fmt"
	"net/http"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
)

const (
	ActionHeader = "SOAPAction"
)

type ActionMux struct {
	handlers map[string]http.Handler
}

func NewActionMux() *ActionMux {
	return &ActionMux{handlers: map[string]http.Handler{}}
}

func (m *ActionMux) Handle(actionUri string, handler http.Handler) *ActionMux {
	_, ok := m.handlers[actionUri]
	if ok {
		panic(errors.Errorf("handler for action %v is already provided", actionUri))
	}
	m.handlers[actionUri] = handler
	return m
}

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
