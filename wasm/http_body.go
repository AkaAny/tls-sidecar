package wasm

import (
	"github.com/pkg/errors"
	"io"
)

func ReadAndClose(respBody io.ReadCloser) ([]byte, error) {
	rawData, err := io.ReadAll(respBody)
	if err != nil {
		return nil, errors.Wrap(err, "read from go reader")
	}
	return rawData, nil
}
