package decoder

import (
	"encoding/json"
	"io"

	"github.com/you/go-web-server/errors"
)

func DecodeJSON(r io.Reader, v interface{}) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return errors.Error(err.Error())
	}
	return nil
}