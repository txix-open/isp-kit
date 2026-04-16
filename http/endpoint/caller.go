package endpoint

import (
	"context"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
)

// param represents a function parameter that can be built from the request context.
type param struct {
	index   int
	builder ParamBuilder
}

// Caller uses reflection to wrap arbitrary functions as HTTP handlers.
// It automatically extracts request body, builds parameters from the context,
// and maps return values to the response.
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

// NewCaller creates a new Caller that wraps the provided function.
// It analyzes the function signature and configures parameter extraction and result mapping.
// Returns an error if the provided value is not a function or has invalid parameter types.
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

// Handle invokes the wrapped function with extracted parameters and maps the result.
// It returns an error if parameter extraction fails or the function returns an error.
func (h *Caller) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	args := make([]reflect.Value, h.paramsCount)

	if h.reqBodyIndex != -1 {
		value, err := h.bodyExtractor.Extract(ctx, r.Body, h.reqBodyType)
		if err != nil {
			return err
		}
		args[h.reqBodyIndex] = value
	}

	for _, p := range h.params {
		value, err := p.builder(ctx, w, r)
		if err != nil {
			return err
		}
		args[p.index] = reflect.ValueOf(value)
	}

	out := h.handler.Call(args)

	if h.hasError && !out[h.errorIndex].IsNil() {
		return out[h.errorIndex].Interface().(error) // nolint:forcetypeassert
	}

	if h.hasResult {
		return h.bodyMapper.Map(ctx, out[0].Interface(), w)
	}

	return nil
}
