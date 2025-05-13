package httpcli

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
)

type MultipartData struct {
	Files  map[string]MultipartFieldFile
	Values map[string]string
}

type MultipartFieldFile struct {
	Headers  map[string]string
	Filename string
	Reader   io.ReadCloser
}

func (m *MultipartData) openReader() (string, io.ReadCloser) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	go func() {
		defer func() {
			_ = writer.Close()
			_ = pw.Close()
			for _, file := range m.Files {
				_ = file.Reader.Close()
			}
		}()
		for key, value := range m.Values {
			err := writer.WriteField(key, value)
			if err != nil {
				return
			}
		}
		for name, file := range m.Files {
			wr, err := CreateFormFile(writer, file.Headers, name, file.Filename)
			if err != nil {
				return
			}
			_, err = io.Copy(wr, file.Reader)
			if err != nil {
				return
			}
		}
	}()
	return writer.FormDataContentType(), pr
}

func CreateFormFile(
	w *multipart.Writer,
	header map[string]string,
	fieldname string,
	filename string,
) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	for key, value := range header {
		h.Set(key, value)
	}
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname), escapeQuotes(filename)))
	return w.CreatePart(h)
}

// nolint:gochecknoglobals
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
