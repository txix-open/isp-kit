package endpoint

import (
	"context"
	"reflect"
	"unicode"

	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/json"
	"github.com/pkg/errors"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Validator interface {
	Validate(value interface{}) (bool, map[string]string)
}

type JsonRequestExtractor struct {
	validator Validator
}

func (j JsonRequestExtractor) Extract(ctx context.Context, message *isp.Message, reqBodyType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(reqBodyType)
	err := json.Unmarshal(message.GetBytesBody(), instance.Interface())
	if err != nil {
		return reflect.Value{}, status.Errorf(codes.InvalidArgument, "unmarshal request body: %v", err)
	}

	elem := instance.Elem()

	ok, details := j.validator.Validate(elem.Interface())
	if ok {
		return elem, nil
	}

	var violations = make([]*epb.BadRequest_FieldViolation, len(details))
	i := 0
	for k, v := range details {
		arr := []rune(k)
		arr[0] = unicode.ToLower(arr[0])
		violations[i] = &epb.BadRequest_FieldViolation{
			Field:       string(arr),
			Description: v,
		}
		i++
	}
	withDetails, err := status.New(codes.InvalidArgument, "invalid request body").
		WithDetails(&epb.BadRequest{FieldViolations: violations})
	if err != nil {
		return reflect.Value{}, errors.WithMessage(err, "make status")
	}
	return reflect.Value{}, withDetails.Err()
}
