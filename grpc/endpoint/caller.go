package endpoint

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/isp"
)

// param represents a single parameter to be injected into a wrapped function.
type param struct {
	index   int
	builder ParamBuilder
}

// Caller wraps a function for gRPC invocation using reflection.
// It analyzes the function signature, injects dependencies, and handles
// request/response marshaling automatically.
type Caller struct {
	bodyExtractor RequestBodyExtractor
	bodyMapper    ResponseBodyMapper

	handler      reflect.Value
	paramsCount  int
	params       []param
	reqBodyIndex int
	reqBodyType  reflect.Type

	hasResult  bool
	hasError   bool
	errorIndex int
}

// NewCaller creates a new Caller for the specified function.
// Analyzes the function signature to determine parameter types and return values.
// Returns an error if the input is not a function or if parameter types cannot be resolved.
func NewCaller(
	f any,
	bodyExtractor RequestBodyExtractor,
	bodyMapper ResponseBodyMapper,
	paramMappers map[string]ParamMapper,
) (*Caller, error) {
	rt := reflect.TypeOf(f)
	if rt.Kind() != reflect.Func {
		return nil, errors.New("function expected")
	}

	paramsCount := rt.NumIn()
	reqBodyIndex := -1
	handler := reflect.ValueOf(f)
	var reqBodyType reflect.Type
	params := make([]param, 0, paramsCount)

	for i := range paramsCount {
		p := rt.In(i)
		paramType := p.String()
		mapper, ok := paramMappers[paramType]

		if !ok { // maybe it's a request body
			if reqBodyIndex != -1 {
				return nil, errors.Errorf("param mapper not found for type %s", paramType)
			}
			reqBodyIndex = i
			reqBodyType = p
			continue
		}

		// it's a simple param
		params = append(params, param{index: i, builder: mapper.Builder})
	}

	numOut := rt.NumOut()
	hasResult := numOut > 0 && rt.Out(0) != reflect.TypeFor[*error]().Elem()
	hasError := numOut > 0 && rt.Out(numOut-1) == reflect.TypeFor[*error]().Elem()

	return &Caller{
		bodyExtractor: bodyExtractor,
		bodyMapper:    bodyMapper,
		handler:       handler,
		paramsCount:   paramsCount,
		params:        params,
		reqBodyIndex:  reqBodyIndex,
		reqBodyType:   reqBodyType,
		hasResult:     hasResult,
		hasError:      hasError,
		errorIndex:    numOut - 1,
	}, nil
}

// Handle executes the wrapped function with the provided context and message.
// Injects parameters, extracts request body, calls the function, and maps the result.
// Returns the response message and any error from the function or parameter extraction.
func (h *Caller) Handle(ctx context.Context, message *isp.Message) (*isp.Message, error) {
	args := make([]reflect.Value, h.paramsCount)

	if h.reqBodyIndex != -1 {
		value, err := h.bodyExtractor.Extract(ctx, message, h.reqBodyType)
		if err != nil {
			return nil, err
		}
		args[h.reqBodyIndex] = value
	}

	for _, p := range h.params {
		value, err := p.builder(ctx, message)
		if err != nil {
			return nil, err
		}
		args[p.index] = reflect.ValueOf(value)
	}

	out := h.handler.Call(args)

	if h.hasError && !out[h.errorIndex].IsNil() {
		return nil, out[h.errorIndex].Interface().(error) // nolint:forcetypeassert
	}

	if h.hasResult {
		return h.bodyMapper.Map(out[0].Interface())
	}

	return nil, nil // nolint:nilnil
}
