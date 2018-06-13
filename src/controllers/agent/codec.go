package agent

import "encoding/json"

type ICodec interface {
	Encode(data interface{}) ([]byte, error)
	Decode(data []byte) (interface{}, error)
}
type Codec struct {}

func (c Codec) Encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
func (c Codec) Decode(data []byte) (interface{}, error) {
	var d interface{}
	err := json.Unmarshal(data, &d)
	return d, err
}
