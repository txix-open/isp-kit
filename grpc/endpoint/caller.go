package endpoint

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/isp"
)

type param struct {
	index   int
	builder ParamBuilder
}

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
	hasResult := numOut > 0 && rt.Out(0) != reflect.TypeOf((*error)(nil)).Elem()
	hasError := numOut > 0 && rt.Out(numOut-1) == reflect.TypeOf((*error)(nil)).Elem()

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
		// Приведение к error безопасно:
		// 1) hasError == true → функция возвращает последним аргументом error
		// 2) проверка errVal != nil исключает typed-nil
		return nil, out[h.errorIndex].Interface().(error) // nolint:forcetypeassert
	}

	if h.hasResult {
		return h.bodyMapper.Map(out[0].Interface())
	}

	return nil, nil // nolint:nilnil
}
