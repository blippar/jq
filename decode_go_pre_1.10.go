// +build !go1.10

package jq

import "encoding/json"

func decodeJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
