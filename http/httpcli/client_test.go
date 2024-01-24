package httpcli_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/json"
	"github.com/integration-system/isp-kit/retry"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

type example struct {
	Data string
}

func TestRequestBuilder_DoWithoutResponse(t *testing.T) {
	require := require.New(t)
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		http.Error(writer, "some error", http.StatusBadRequest)
	})).URL
	err := httpcli.New().Get(url).StatusCodeToError().DoWithoutResponse(context.Background())
	require.Error(err)
	httpErr := httpcli.ErrorResponse{}
	ok := errors.As(err, &httpErr)
	require.True(ok)
	require.EqualValues("some error\n", string(httpErr.Body))
	require.EqualValues(http.StatusBadRequest, httpErr.StatusCode)
	t.Log(httpErr.Error())
}

func TestRequestBuilder_Header(t *testing.T) {
	require := require.New(t)
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		require.EqualValues("123", r.Header.Get("x-header"))
		writer.WriteHeader(http.StatusOK)
	})).URL
	resp, err := httpcli.New().Get(url).Header("x-header", "123").Do(context.Background())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_Cookie(t *testing.T) {
	require := require.New(t)
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		var actual *http.Cookie
		for i := range r.Cookies() {
			cookie := r.Cookies()[i]
			if cookie.Name == "x-cookie" {
				actual = cookie
			}
		}
		require.EqualValues("123", actual.Value)
		writer.WriteHeader(http.StatusOK)
	})).URL
	resp, err := httpcli.New().
		Get(url).
		Cookie(&http.Cookie{Name: "x-cookie", Value: "123"}).
		Do(context.Background())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_FormDataRequestBody(t *testing.T) {
	require := require.New(t)
	expected := map[string][]string{
		"1":    {"1", "2", "3"},
		"test": {"value"},
	}
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(err)
		require.EqualValues(expected, r.PostForm)
		writer.WriteHeader(http.StatusOK)
	})).URL
	resp, err := httpcli.New().Post(url).FormDataRequestBody(expected).Do(context.Background())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_BasicAuth(t *testing.T) {
	require := require.New(t)
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		require.True(ok)
		require.EqualValues("user", username)
		require.EqualValues("pass", password)
		writer.WriteHeader(http.StatusOK)
	})).URL
	resp, err := httpcli.New().Get(url).
		BasicAuth(httpcli.BasicAuth{
			Username: "user",
			Password: "pass",
		}).Do(context.Background())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_QueryParams(t *testing.T) {
	require := require.New(t)
	request := map[string]any{
		"1":    "2",
		"test": 1,
	}
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(err)
		expected := url.Values{
			"1":    {"2"},
			"test": {"1"},
		}
		require.EqualValues(expected, r.Form)
		writer.WriteHeader(http.StatusOK)
	})).URL
	resp, err := httpcli.New().Get(url).QueryParams(request).Do(context.Background())
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestRequestBuilder_Retry(t *testing.T) {
	require := require.New(t)
	callCount := 0
	middlewareCallCount := 0
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 3 {
			data, err := io.ReadAll(r.Body)
			require.NoError(err)
			_, err = writer.Write(data)
			require.NoError(err)
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
		Do(context.Background())
	require.EqualValues(3, callCount)
	require.EqualValues(1, middlewareCallCount)
	require.NoError(err)
	require.True(resp.IsSuccess())
	require.EqualValues("test_data", exm.Data)
}

func TestRequestBuilder_RetryOnError(t *testing.T) {
	require := require.New(t)
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	srv.Close()
	cli := httpcli.New()
	resp, err := cli.Get(srv.URL).
		Retry(httpcli.IfErrorOr5XXStatus(), retry.NewExponentialBackoff(3*time.Second)).
		JsonRequestBody(example{Data: "test_data"}).
		Do(context.Background())
	require.Error(err)
	require.Nil(resp)
}

func TestRequestBuilder_RetryWithTimeout(t *testing.T) {
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
		Do(context.Background())
	t.Log(err)
	require.Error(err)
	require.Nil(resp)
}

func TestRequestBuilder_GlobalRequestTimeout(t *testing.T) {
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
		Do(context.Background())
	t.Log(err)
	require.Error(err)
	require.Nil(resp)
}

func TestRequestBuilder_MultipartData(t *testing.T) {
	require := require.New(t)
	url := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(1024)
		require.NoError(err)
		defer func() { _ = r.MultipartForm.RemoveAll() }()

		require.EqualValues("test", r.MultipartForm.Value["field1"][0])

		fh := r.MultipartForm.File["file"][0]
		file, err := fh.Open()
		require.NoError(err)
		defer file.Close()
		data, err := io.ReadAll(file)
		require.NoError(err)
		_, err = writer.Write(data)
		require.NoError(err)
	})).URL

	file, err := os.Open("test_data/multipart.json")
	require.NoError(err)
	resp, err := httpcli.New().Post(url).MultipartRequestBody(&httpcli.MultipartData{
		Files:  map[string]httpcli.MultipartFieldFile{"file": {Filename: "multipart.json", Reader: file}},
		Values: map[string]string{"field1": "test"},
	}).Do(context.Background())
	require.NoError(err)
	require.True(resp.IsSuccess())

	_, err = file.Read([]byte{0})
	require.Error(err) //file closed

	expected, err := os.ReadFile("test_data/multipart.json")
	require.NoError(err)
	actual, err := resp.Body()
	require.NoError(err)
	require.Equal(expected, actual)
}

func TestConcurrency(t *testing.T) {
	srv := echoServer()
	cli := httpcli.New()
	group, _ := errgroup.WithContext(context.Background())
	group.SetLimit(100)
	for i := 0; i < 10000; i++ {
		group.Go(func() error {
			data := make([]byte, 4200)
			_, _ = rand.Read(data)
			requestData := hex.EncodeToString(data)
			response := example{}
			resp, err := cli.Post(srv.URL).
				JsonRequestBody(example{Data: requestData}).
				JsonResponseBody(&response).
				Do(context.Background())
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
				Do(context.Background())
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
				SetContext(context.Background()).
				Post(srv.URL)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func echoServer() *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		data, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		_, err = w.Write(data)
		if err != nil {
			panic(err)
		}
	}))
	return srv
}
