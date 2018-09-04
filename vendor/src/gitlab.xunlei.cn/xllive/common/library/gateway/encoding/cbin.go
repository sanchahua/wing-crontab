package encoding

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Encoder struct{}
type Decoder struct{}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

//编码方式total_length | [item_length | key_length | key | value_length | value ]...
func (Encoder) Encode(data map[string]interface{}) ([]byte, error) {
	resultBuffer := new(bytes.Buffer)
	for key := range data {
		var str string
		switch t := data[key].(type) {
		case string:
		case float32:
		case float64:
		case int:
		case uint:
		case int8:
		case uint8:
		case int16:
		case uint16:
		case int32:
		case uint32:
		case int64:
		case uint64:
		default:
			return nil, errors.New(fmt.Sprintf("key:%s invalid type:%v", key, t))
		}
		str = fmt.Sprintf("%v", data[key])
		var keyLen uint32 = uint32(len(key))
		var valueLen uint32 = uint32(len(str))
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.BigEndian, keyLen); err != nil {
			return nil, err
		}
		if n, err := buf.Write([]byte(key)); err != nil || n < int(keyLen) {
			return nil, errors.New(fmt.Sprintf("key:%s trans to byte error：%v", key, err))
		}
		if err := binary.Write(buf, binary.BigEndian, valueLen); err != nil {
			return nil, err
		}
		if n, err := buf.Write([]byte(str)); err != nil || n < int(valueLen) {
			return nil, errors.New(fmt.Sprintf("value:%s trans to byte error：%v", str, err))
		}
		var itemLen uint32 = uint32(buf.Len())
		if err := binary.Write(resultBuffer, binary.BigEndian, itemLen); err != nil {
			return nil, err
		}
		if n, err := buf.WriteTo(resultBuffer); err != nil || n < int64(itemLen) {
			return nil, errors.New(fmt.Sprintf("item trans to byte error：%v", err))
		}
		//fmt.Println(keyLen, valueLen, itemLen, resultBuffer.Len())
	}
	var dataLen uint32 = uint32(resultBuffer.Len())
	result := new(bytes.Buffer)
	if err := binary.Write(result, binary.BigEndian, dataLen+8); err != nil {
		return nil, err
	}
	if err := binary.Write(result, binary.BigEndian, dataLen); err != nil {
		return nil, err
	}
	if n, err := resultBuffer.WriteTo(result); err != nil || n < int64(dataLen) {
		return nil, errors.New(fmt.Sprintf("data trans to byte error：%v", err))
	}
	return []byte(result.String()), nil
}

func (Decoder) Decode(reader io.Reader) (map[string]string, error) {
	result := make(map[string]string, 0)
	//reader := bytes.NewReader(data)
	var total uint32 = 0
	if err := binary.Read(reader, binary.BigEndian, &total); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &total); err != nil {
		return nil, err
	}
	content := make([]byte, total)
	if n, err := reader.Read(content); err != nil || n != int(total) {
		return nil, errors.New(fmt.Sprintf("data content read error：%v", err))
	}
	contentReader := bytes.NewReader(content)

	for {
		var itemLen uint32
		var keyLen uint32
		var valueLen uint32
		if err := binary.Read(contentReader, binary.BigEndian, &itemLen); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if err := binary.Read(contentReader, binary.BigEndian, &keyLen); err != nil {
			return nil, err
		}
		key := make([]byte, keyLen)
		if n, err := contentReader.Read(key); err != nil || n != int(keyLen) {
			return nil, errors.New(fmt.Sprintf("data decode to map error：%v", err))
		}
		if err := binary.Read(contentReader, binary.BigEndian, &valueLen); err != nil {
			return nil, err
		}
		value := make([]byte, valueLen)
		if n, err := contentReader.Read(value); err != nil || n != int(valueLen) {
			return nil, errors.New(fmt.Sprintf("data decode to map error：%v", err))
		}
		result[string(key)] = string(value)
	}
	return result, nil
}
