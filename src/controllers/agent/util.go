package agent

import (
	"encoding/binary"
	"errors"
)

var commandLenError = errors.New("command len error")
func unpack(data []byte) (id int64, isMutex byte, command string, dispatchServer string, err error) {
	if len(data) < 25 {
		err = commandLenError
		return
	}
	err          = nil
	id           = int64(binary.LittleEndian.Uint64(data[:8]))
	//dispatchTime = int64(binary.LittleEndian.Uint64(data[8:16]))
	isMutex      = data[8]
	commandLen  := binary.LittleEndian.Uint64(data[9:17])
	if len(data) < int(17 + commandLen) {
		err = commandLenError
		return
	}
	command        = string(data[17:17+commandLen])
	dispatchServer = string(data[17+commandLen:])
	return
}

func pack(item *runItem, bindAddress string) []byte {
	//json.Marshal(item)
	sendData := make([]byte, 8)
	binary.LittleEndian.PutUint64(sendData, uint64(item.id))

	dataCommendLen := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataCommendLen, uint64(len(item.command)))

	//currentTime := make([]byte, 8)
	//binary.LittleEndian.PutUint64(currentTime, uint64(time.Now().Unix()))
	//sendData = append(sendData, currentTime...)

	if item.isMutex {
		sendData = append(sendData, byte(1))
	} else {
		sendData = append(sendData, byte(0))
	}

	sendData = append(sendData, dataCommendLen...)
	sendData = append(sendData, []byte(item.command)...)

	sendData = append(sendData, []byte(bindAddress)...)//c.ctx.Config.BindAddress)...)
	return sendData
}


