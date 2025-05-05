// nolint:gosec,wrapcheck
package httpcli_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	rand2 "math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/retry"
	"golang.org/x/sync/errgroup"
)

type example struct {
	Data string
}

func TestRequestBuilder_DoWithoutResponse(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	url := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, r *http.Request) {
			http.Error(writer, "some error", http.StatusBadRequest)
		},
	)).URL
	err := httpcli.New().Get(url).StatusCodeToError().DoWithoutResponse(t.Context())
	require.Error(err)
	httpErr := httpcli.ErrorResponse{}
	ok := errors.As(err, &httpErr)
	require.True(ok)
	require.EqualValues("some error\n", string(httpErr.Body))
	require.EqualValues(http.StatusBadRequest, httpErr.StatusCode)
	t.Log(httpErr.Error())
}

func TestRequestBuilder_Header(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		assert.EqualValues(t, "123", r.Header.Get("X-Header"))
		writer.WriteHeader(http.StatusOK)
	})).URL
	resp, err := httpcli.New().Get(url).Header("x-header", "123").Do(t.Context())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_Cookie(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	url := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, r *http.Request) {
			var actual *http.Cookie
			for i := range r.Cookies() {
				cookie := r.Cookies()[i]
				if cookie.Name == "x-cookie" {
					actual = cookie
				}
			}
			assert.EqualValues(t, "123", actual.Value)
			writer.WriteHeader(http.StatusOK)
		},
	)).URL
	resp, err := httpcli.New().
		Get(url).
		Cookie(&http.Cookie{Name: "x-cookie", Value: "123"}).
		Do(t.Context())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_FormDataRequestBody(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)
	expected := map[string][]string{
		"1":    {"1", "2", "3"},
		"test": {"value"},
	}
	url := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			assert.NoError(err)
			assert.EqualValues(expected, r.PostForm)
			writer.WriteHeader(http.StatusOK)
		},
	)).URL
	resp, err := httpcli.New().Post(url).FormDataRequestBody(expected).Do(t.Context())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_BasicAuth(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)

	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.True(ok)
		assert.EqualValues("user", username)
		assert.EqualValues("pass", password)
		writer.WriteHeader(http.StatusOK)
	})).URL
	resp, err := httpcli.New().Get(url).
		BasicAuth(httpcli.BasicAuth{
			Username: "user",
			Password: "pass",
		}).Do(t.Context())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_QueryParams(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)

	request := map[string]any{
		"1":    "2",
		"test": 1,
	}
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		assert.NoError(err)
		expected := url.Values{
			"1":    {"2"},
			"test": {"1"},
		}
		assert.EqualValues(expected, r.Form)
		writer.WriteHeader(http.StatusOK)
	})).URL
	resp, err := httpcli.New().Get(url).QueryParams(request).Do(t.Context())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_Retry(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)

	callCount := 0
	middlewareCallCount := 0
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 3 {
			data, err := io.ReadAll(r.Body)
			assert.NoError(err)
			_, err = writer.Write(data)
			assert.NoError(err)
			return
		}
		writer.WriteHeader(http.StatusInternalServerError)
	})).URL
	middleware := func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			middlewareCallCount++
			return next.RoundTrip(ctx, request)
		})
	}
	exm := example{}
	cli := httpcli.New(httpcli.WithMiddlewares(middleware))

	resp, err := cli.
		Post(url).
		Retry(httpcli.IfErrorOr5XXStatus(), retry.NewExponentialBackoff(5*time.Second)).
		JsonRequestBody(example{Data: "test_data"}).
		JsonResponseBody(&exm).
		Do(t.Context())
	require.EqualValues(3, callCount)
	require.EqualValues(1, middlewareCallCount)
	require.NoError(err)
	require.True(resp.IsSuccess())
	require.EqualValues("test_data", exm.Data)
}

func TestRequestBuilder_RetryOnError(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	srv.Close()
	cli := httpcli.New()
	resp, err := cli.Get(srv.URL).
		Retry(httpcli.IfErrorOr5XXStatus(), retry.NewExponentialBackoff(3*time.Second)).
		JsonRequestBody(example{Data: "test_data"}).
		Do(t.Context())
	require.Error(err)
	require.Nil(resp)
}

func TestRequestBuilder_RetryWithTimeout(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		t.Log("call endpoint")
		time.Sleep(500 * time.Millisecond)
		writer.WriteHeader(http.StatusOK)
	}))
	cli := httpcli.New()
	resp, err := cli.Get(srv.URL).
		Retry(httpcli.IfErrorOr5XXStatus(), retry.NewExponentialBackoff(1*time.Second)).
		Timeout(100 * time.Millisecond).
		JsonRequestBody(example{Data: "test_data"}).
		Do(t.Context())
	t.Log(err)
	require.Error(err)
	require.Nil(resp)
}

func TestRequestBuilder_GlobalRequestTimeout(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		t.Log("call endpoint")
		time.Sleep(500 * time.Millisecond)
		writer.WriteHeader(http.StatusOK)
	}))
	cli := httpcli.New()
	cli.GlobalRequestConfig().Timeout = 100 * time.Millisecond
	resp, err := cli.Get(srv.URL).
		Retry(httpcli.IfErrorOr5XXStatus(), retry.NewExponentialBackoff(2*time.Second)).
		JsonRequestBody(example{Data: "test_data"}).
		Do(t.Context())
	t.Log(err)
	require.Error(err)
	require.Nil(resp)
}

func TestRequestBuilder_MultipartData(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)

	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(1024)
		assert.NoError(err)
		defer func() { _ = r.MultipartForm.RemoveAll() }()

		assert.EqualValues("test", r.MultipartForm.Value["field1"][0])

		fh := r.MultipartForm.File["file"][0]
		contentTypeHeader := fh.Header.Get("Content-Type")
		assert.EqualValues("application/json1", contentTypeHeader)
		file, err := fh.Open()
		assert.NoError(err)
		defer file.Close()
		data, err := io.ReadAll(file)
		assert.NoError(err)
		_, err = writer.Write(data)
		assert.NoError(err)
	})).URL

	file, err := os.Open("test_data/multipart.json")
	require.NoError(err)
	resp, err := httpcli.New().Post(url).MultipartRequestBody(&httpcli.MultipartData{
		Files: map[string]httpcli.MultipartFieldFile{"file": {
			Filename: "multipart.json",
			Reader:   file,
			Headers: map[string]string{
				"Content-Type": "application/json1",
			},
		}},
		Values: map[string]string{"field1": "test"},
	}).Do(t.Context())
	require.NoError(err)
	require.True(resp.IsSuccess())

	_, err = file.Read([]byte{0})
	require.Error(err) // file closed

	expected, err := os.ReadFile("test_data/multipart.json")
	require.NoError(err)
	actual, err := resp.Body()
	require.NoError(err)
	require.Equal(expected, actual)
}

func TestRequestBuilder_Middlewares(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)

	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		assert.EqualValues("5", r.Header.Get("Content-Length"))
		assert.NotEqual("chunked", r.Header.Get("Transfer-Encoding"))
		writer.WriteHeader(http.StatusOK)
	})).URL
	err := httpcli.New().Post(url).
		RequestBody([]byte("hello")).
		StatusCodeToError().
		Middlewares(httpcli.SetContentLength()).
		DoWithoutResponse(t.Context())
	require.NoError(err)
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	srv := echoServer()
	cli := httpcli.New()
	group, _ := errgroup.WithContext(t.Context())
	group.SetLimit(100)
	for range 10000 {
		group.Go(func() error {
			data := make([]byte, 4200)
			_, _ = rand.Read(data)
			requestData := hex.EncodeToString(data)
			response := example{}
			resp, err := cli.Post(srv.URL).
				JsonRequestBody(example{Data: requestData}).
				JsonResponseBody(&response).
				Retry(func(err error, response *httpcli.Response) error {
					if err != nil {
						return err
					}
					_, err = response.Body()
					if err != nil {
						return err
					}
					if rand2.Int()%2 == 0 {
						return errors.New("retry")
					}
					return nil
				}, retry.NewExponentialBackoff(500*time.Millisecond)).
				Do(t.Context())
			if err != nil {
				return err
			}
			defer resp.Close()
			if response.Data != requestData {
				t.Fatal(response.Data, "not equal", requestData)
			}
			return nil
		})
	}
	err := group.Wait()
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkClient_Post(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		srv := echoServer()
		cli := httpcli.New()
		cli.GlobalRequestConfig().Timeout = 0
		for pb.Next() {
			data := make([]byte, 4200)
			_, _ = rand.Read(data)
			resp, err := cli.Post(srv.URL).
				JsonRequestBody(example{Data: hex.EncodeToString(data)}).
				JsonResponseBody(&example{}).
				Do(b.Context())
			if err != nil {
				b.Fatal(err)
			}
			resp.Close()
		}
	})
}

func BenchmarkGoResty(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		srv := echoServer()
		cli := resty.New()
		cli.JSONMarshal = json.Marshal
		cli.JSONUnmarshal = json.Unmarshal
		for pb.Next() {
			data := make([]byte, 4200)
			_, _ = rand.Read(data)
			_, err := cli.R().
				SetBody(example{Data: hex.EncodeToString(data)}).
				SetResult(&example{}).
				SetContext(b.Context()).
				Post(srv.URL)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func echoServer() *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			data, err := io.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			if !jsoniter.Valid(data) {
				panic(errors.Errorf("invalid json: %s", data))
			}
			_, err = w.Write(data)
			if err != nil {
				panic(err)
			}
		},
	))
	return srv
}
