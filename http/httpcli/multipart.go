package httpcli

import (
	"io"
	"mime/multipart"
)

type MultipartData struct {
	Files  map[string]MultipartFieldFile
	Values map[string]string
}

type MultipartFieldFile struct {
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
			wr, err := writer.CreateFormFile(name, file.Filename)
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
