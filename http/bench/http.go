package main

import (
	"context"
	"io"
	"net"
	stdHttp "net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/integration-system/isp-kit/http"
	"github.com/integration-system/isp-kit/http/endpoint"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/labstack/echo/v4"
)

type Request struct {
	S string `validate:"required"`
}

type Response struct {
	FromReq string
	Resp    string
}

func main() {
	ServeIspHttp()
	ServeEchoHttp()
	ServeGinHttp()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

func ServeIspHttp() {
	logger, err := log.New()
	if err != nil {
		panic(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		panic(err)
	}
	s := http.NewServer()
	go func() {
		err := s.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	handler := func(ctx context.Context, req Request) (*Response, error) {
		return &Response{
			FromReq: req.S,
			Resp:    uuid.New().String(),
		}, nil
	}
	httpHandler := endpoint.DefaultWrapper(logger).Endpoint(handler)
	mux := httprouter.New()
	mux.Handler(stdHttp.MethodPost, "/post", httpHandler)
	s.Upgrade(mux)
}

type v struct {
}

func (v v) Validate(i any) error {
	return validator.Default.ValidateToError(i)
}

func ServeEchoHttp() {
	listener, err := net.Listen("tcp", "127.0.0.1:8001")
	if err != nil {
		panic(err)
	}
	e := echo.New()
	e.Validator = v{}
	e.Listener = listener
	e.HideBanner = true
	e.Logger.SetOutput(io.Discard)

	e.POST("/post", func(c echo.Context) error {
		req := Request{}
		err := c.Bind(&req)
		if err != nil {
			return err
		}
		err = c.Validate(req)
		if err != nil {
			return err
		}
		err = c.JSON(stdHttp.StatusOK, Response{
			FromReq: req.S,
			Resp:    uuid.NewString(),
		})
		if err != nil {
			return err
		}
		return nil
	})
	go func() {
		err := e.Start("")
		if err != nil {
			panic(err)
		}
	}()
}

func ServeGinHttp() {
	listener, err := net.Listen("tcp", "127.0.0.1:8002")
	if err != nil {
		panic(err)
	}
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.POST("/post", func(c *gin.Context) {
		req := Request{}
		err := c.BindJSON(&req)
		if err != nil {
			_ = c.AbortWithError(stdHttp.StatusBadRequest, err)
		}
		err = validator.Default.ValidateToError(req)
		if err != nil {
			_ = c.AbortWithError(stdHttp.StatusBadRequest, err)
		}
		c.JSON(stdHttp.StatusOK, Response{
			FromReq: req.S,
			Resp:    uuid.NewString(),
		})
	})

	go func() {
		err := g.RunListener(listener)
		if err != nil {
			panic(err)
		}
	}()
}
