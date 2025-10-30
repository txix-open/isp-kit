package endpoint_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/http/router"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/fake"
	"github.com/txix-open/isp-kit/test/httpt"
)

func TestEndpoint(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(endpointTestSuite))
}

type endpointTestSuite struct {
	suite.Suite

	test               *test.Test
	apiCli             *httpcli.Client
	noRespCounter      *atomic.Int64
	withHttpReqCounter *atomic.Int64
}

func (s *endpointTestSuite) SetupTest() {
	s.test, _ = test.New(s.T())
	w := endpoint.DefaultWrapper(s.test.Logger(), httplog.Noop())

	s.noRespCounter = new(atomic.Int64)
	s.withHttpReqCounter = new(atomic.Int64)
	c := newController(s.noRespCounter, s.withHttpReqCounter)

	r := router.New()
	r.POST("/basic", endpoint.New(c.Basic).Wrap(w))
	r.POST("/without-resp", endpoint.NewWithoutResponse(c.WithoutResponse).Wrap(w))
	r.POST("/with-http-req", endpoint.NewWithRequest(c.WithHttpRequest).Wrap(w))
	r.POST("/default", endpoint.NewDefaultHttp(c.DefaultHttp).Wrap(w))

	_, s.apiCli = httpt.TestServer(s.test, r, httpcli.WithMiddlewares(httpclix.Log(s.test.Logger())))
}

func (s *endpointTestSuite) TestEndpoint_Basic() {
	var (
		req  = fake.It[request]()
		resp = new(response)
	)
	err := s.apiCli.Post("/basic").
		JsonRequestBody(req).
		JsonResponseBody(resp).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)
	s.Require().Equal(req.Input, resp.Output)

	err = s.apiCli.Post("/basic").
		JsonRequestBody(request{Input: ""}).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	s.Require().NotNil(err) // nolint:testifylint
}

func (s *endpointTestSuite) TestEndpoint_WithHttpRequest() {
	s.Require().Zero(s.withHttpReqCounter.Load())
	err := s.apiCli.Post("/with-http-req").
		JsonRequestBody(fake.It[request]()).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)
	s.Require().EqualValues(1, s.withHttpReqCounter.Load())
}

func (s *endpointTestSuite) TestEndpoint_WithoutResponseBody() {
	s.Require().Zero(s.noRespCounter.Load())
	err := s.apiCli.Post("/without-resp").
		JsonRequestBody(fake.It[request]()).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)
	s.Require().EqualValues(1, s.noRespCounter.Load())
}

func (s *endpointTestSuite) TestEndpoint_DefaultHttp() {
	var (
		req  = fake.It[request]()
		resp = new(response)
	)
	err := s.apiCli.Post("/default").
		JsonRequestBody(req).
		JsonResponseBody(resp).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)
	s.Require().Equal(req.Input, resp.Output)
}

type request struct {
	Input string `validate:"required"`
}

type response struct {
	Output string
}

type controller struct {
	noRespCounter      *atomic.Int64
	withHttpReqCounter *atomic.Int64
}

func newController(noRespCounter *atomic.Int64, withHttpReqCounter *atomic.Int64) controller {
	return controller{
		noRespCounter:      noRespCounter,
		withHttpReqCounter: withHttpReqCounter,
	}
}

func (c *controller) Basic(ctx context.Context, req request) (response, error) {
	return response{Output: req.Input}, nil
}

func (c *controller) WithoutResponse(ctx context.Context, req request) error {
	c.noRespCounter.Add(1)
	return nil
}

func (c *controller) WithHttpRequest(ctx context.Context, r *http.Request) error {
	c.withHttpReqCounter.Add(1)
	return nil
}

func (c *controller) DefaultHttp(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	req := new(request)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(response{Output: req.Input})
}
