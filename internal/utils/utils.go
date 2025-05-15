package utils

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

func CopyBuffer(dst http.ResponseWriter, src io.Reader, buf []byte) ([]byte, error) {
	var res bytes.Buffer

	if len(buf) == 0 {
		buf = make([]byte, 32*1024)
	}
	var written int64
	for {
		nr, rerr := src.Read(buf)
		if rerr != nil && rerr != io.EOF && rerr != context.Canceled {
			return res.Bytes(), rerr
		}
		if nr > 0 {
			nw, werr := dst.Write(buf[:nr])
			res.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if werr != nil {
				return res.Bytes(), werr
			}
			if nr != nw {
				return res.Bytes(), io.ErrShortWrite
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				rerr = nil
			}
			return res.Bytes(), rerr
		}
		http.NewResponseController(dst).Flush()
	}
}
