package codec

import "io"

type readCloser struct {
	decoder decoder
	pool    putter
	body    io.Closer
}

func (rc *readCloser) Read(p []byte) (n int, err error) {
	return rc.decoder.Read(p)
}

func (rc *readCloser) Close() error {
	// reset to avoid leaking state between uses
	_ = rc.decoder.Reset(nil)

	// return to pool
	rc.pool.Put(rc.decoder)
	rc.decoder = nil

	return rc.body.Close()
}

type writeCloser struct {
	encoder encoder
	pool    putter
}

func (wc *writeCloser) Write(p []byte) (int, error) {
	return wc.encoder.Write(p)
}

func (wc *writeCloser) Close() error {
	err := wc.encoder.Close()

	// reset to avoid leaking state between uses
	wc.encoder.Reset(nil)

	// return to pool
	wc.pool.Put(wc.encoder)
	wc.encoder = nil

	return err
}
