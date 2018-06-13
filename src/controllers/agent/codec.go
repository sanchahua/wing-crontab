package agent

import "encoding/json"

// 基本的数据包编码解码实现
type ICodec interface {
	Encode(event int, data interface{}) ([]byte, error)
	Decode(data []byte) (int, interface{}, error)
}
// 发送数据包的基本构成
// Event代表对应的事件
// Data对应基本的数据
type Package struct {
	Event int
	Data interface{}
}
type Codec struct {}

func (c Codec) Encode(event int, data interface{}) ([]byte, error) {
	p := Package{event, data}
	return json.Marshal(p)
}

func (c Codec) Decode(data []byte) (int, interface{}, error) {
	var d Package
	err := json.Unmarshal(data, &d)
	return d.Event, d.Data, err
}
