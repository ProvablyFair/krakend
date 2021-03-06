package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/google/martian/parse"
)

func init() {
	parse.Register("provablyFair.BackendSelector", backendSelectorFromJSON)
}

// BackendSelector substitutes the host of the HTTP request depending on the value of
// the `Game` field of the requesy body
type BackendSelector struct {
	Mapping map[string]string
}

// Err404 is the error returned by the modifier if the requested game is not defined
// in the mapping injected to the BackendSelector
var Err404 = errors.New("unnknown game")

// ModifyRequest extracts the value of the `Game` field from the request and updates the
// Host of the request URL using in the injected mapping
func (b *BackendSelector) ModifyRequest(req *http.Request) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	s := new(Session)
	if err := json.Unmarshal(body, s); err != nil {
		return err
	}

	backend, ok := b.Mapping[s.Game]
	if !ok {
		return Err404
	}

	req.URL.Host = backend
	return nil
}

// Session is struct for quick param extraction from the request body
type Session struct {
	Game string `json:"game"`
}

// backendSelectorFromJSON is the factory used by the martian package to instantiate
// the modifier registered under the name "provablyFair.BackendSelector"
func backendSelectorFromJSON(b []byte) (*parse.Result, error) {
	mapping := map[string]string{}

	if err := json.Unmarshal(b, &mapping); err != nil {
		return nil, err
	}

	return parse.NewResult(&BackendSelector{Mapping: mapping}, []parse.ModifierType{parse.Request})
}
