package grpc_test

import (
	"context"
	"net"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/apierrors"
	grpcCli "github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type reqBody struct {
	A string
	B bool
	C int32
}

type respBody struct {
	Ok bool
}

func TestGrpcBasic(t *testing.T) {
	t.Parallel()

	require, srv, cli := prepareTest(t)
	reqId := requestid.Next()
	ctx := requestid.ToContext(t.Context(), reqId)
	expectedReq := reqBody{
		A: "string",
		B: true,
		C: 123,
	}
	handler := func(ctx context.Context, data grpc.AuthData, req reqBody) (*respBody, error) {
		receivedReqId := requestid.FromContext(ctx)
		require.EqualValues(reqId, receivedReqId)

		appId, err := data.ApplicationId()
		require.NoError(err)
		require.EqualValues(123, appId)

		require.EqualValues(expectedReq, req)

		return &respBody{Ok: true}, nil
	}
	logger, err := log.New()
	require.NoError(err)
	wrapper := endpoint.DefaultWrapper(logger)
	srv.Upgrade(grpc.NewMux().Handle("endpoint", wrapper.Endpoint(handler)))

	resp := respBody{}

	err = cli.Invoke("endpoint").
		ApplicationId(123).
		JsonRequestBody(expectedReq).
		JsonResponseBody(&resp).
		Do(ctx)
	require.NoError(err)
	require.True(resp.Ok)
}

func TestGrpcValidation(t *testing.T) {
	t.Parallel()

	require, srv, cli := prepareTest(t)

	type reqBody struct {
		A string `validate:"required"`
	}
	logger, err := log.New()
	require.NoError(err)
	wrapper := endpoint.DefaultWrapper(logger)
	callCount := int32(0)
	handler := grpc.NewMux().Handle("endpoint", wrapper.Endpoint(func(req reqBody) {
		atomic.AddInt32(&callCount, 1)
	}))
	srv.Upgrade(handler)

	err = cli.Invoke("endpoint").JsonRequestBody(reqBody{A: ""}).Do(t.Context())
	require.EqualValues(codes.InvalidArgument, status.Code(err))
	apiError := apierrors.FromError(err)
	require.NotNil(apiError)
	require.EqualValues(apierrors.Error{
		ErrorCode:    400,
		ErrorMessage: "invalid request body",
		Details: map[string]any{
			"a": "A is a required field",
		},
	}, *apiError)
	require.EqualValues(0, atomic.LoadInt32(&callCount))
}

func TestRequestIdChain(t *testing.T) {
	t.Parallel()

	_, srv1, cli1 := prepareTest(t)
	require, srv2, cli2 := prepareTest(t)

	reqId := requestid.Next()
	ctx := requestid.ToContext(t.Context(), reqId)
	callCount := int32(0)
	handler1 := func(ctx context.Context) {
		receivedReqId := requestid.FromContext(ctx)
		require.EqualValues(reqId, receivedReqId)
		err := cli2.Invoke("endpoint2").Do(ctx)
		require.NoError(err)
	}
	handler2 := func(ctx context.Context) {
		receivedReqId := requestid.FromContext(ctx)
		require.EqualValues(reqId, receivedReqId)

		atomic.AddInt32(&callCount, 1)
	}
	logger, err := log.New()
	require.NoError(err)
	wrapper := endpoint.DefaultWrapper(logger)
	srv1.Upgrade(grpc.NewMux().Handle("endpoint1", wrapper.Endpoint(handler1)))
	srv2.Upgrade(grpc.NewMux().Handle("endpoint2", wrapper.Endpoint(handler2)))

	err = cli1.Invoke("endpoint1").Do(ctx)
	require.NoError(err)
	require.EqualValues(1, atomic.LoadInt32(&callCount))
}

func TestGrpcAppendMetadata(t *testing.T) {
	t.Parallel()

	require, srv, cli := prepareTest(t)
	reqId := requestid.Next()
	testKey := "test-key"
	testValue := "testValue"
	ctx := requestid.ToContext(t.Context(), reqId)

	handler := func(ctx context.Context, data grpc.AuthData) (*respBody, error) {
		receivedReqId := requestid.FromContext(ctx)
		require.EqualValues(reqId, receivedReqId)

		appId, err := data.ApplicationId()
		require.NoError(err)
		require.EqualValues(123, appId)

		value, err := grpc.StringFromMd(testKey, metadata.MD(data))
		require.NoError(err)
		require.Equal(testValue, value)

		return &respBody{Ok: true}, nil
	}
	logger, err := log.New()
	require.NoError(err)
	wrapper := endpoint.DefaultWrapper(logger)
	srv.Upgrade(grpc.NewMux().Handle("endpoint", wrapper.Endpoint(handler)))

	resp := respBody{}

	err = cli.Invoke("endpoint").
		ApplicationId(123).
		AppendMetadata(testKey, testValue).
		JsonResponseBody(&resp).
		Do(ctx)
	require.NoError(err)
	require.True(resp.Ok)
}

func prepareTest(t *testing.T) (*require.Assertions, *grpc.Server, *grpcCli.Client) {
	t.Helper()
	required := require.New(t)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	required.NoError(err)
	srv := grpc.NewServer()
	cli, err := grpcCli.Default()
	required.NoError(err)
	t.Cleanup(func() {
		err := cli.Close()
		required.NoError(err)
		srv.Shutdown()
	})
	go func() {
		err := srv.Serve(listener)
		assert.NoError(t, err)
	}()

	cli.Upgrade([]string{listener.Addr().String()})
	return required, srv, cli
}
