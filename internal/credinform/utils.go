package credinform

import "encoding/json"

func DecodeToType(input interface{}, out interface{}) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
