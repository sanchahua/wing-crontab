package agent

import (
	"errors"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
)

func hasCmd(cmd int) bool {
	return cmd == CMD_ERROR||
		cmd == CMD_TICK ||
		cmd == CMD_AGENT||
		cmd == CMD_STOP||
		cmd == CMD_RELOAD||
		cmd == CMD_SHOW_MEMBERS||
		cmd == CMD_CRONTAB_CHANGE ||
		cmd == CMD_RUN_COMMAND ||
		cmd == CMD_PULL_COMMAND ||
		cmd == CMD_DEL_CACHE
}

func Pack(cmd int, msg []byte) []byte {
	//start := time.Now()
	//defer log.Debugf("Pack use time %+v", time.Since(start))
	l  := len(msg)
	r  := make([]byte, l+6)
	cl := l + 2
	binary.LittleEndian.PutUint32(r[:4], uint32(cl))
	//log.Debugf("pack cl=%+v", r[:4])
	binary.LittleEndian.PutUint16(r[4:6], uint16(cmd))
	copy(r[6:], msg)
	//log.Debugf("pack (cmd=%v)(msg=%v) == %+v", cmd, msg, r)
	return r
}

var DataLenError = errors.New("data len error")
var MaxPackError = errors.New("package len error")
// return cmd, content, endPoint, error
func Unpack(data *[]byte) (int, []byte, error) {

	if data == nil || len(*data) == 0 {
		return 0, nil, nil
	}
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Unpack recover##########%+v", err)
			*data = make([]byte, 0)
		}
	}()
	//log.Debugf("data: %+v", *data)
	if len(*data) > MAX_PACKAGE_LEN {
		log.Errorf("max len error: %+v", *data)
		*data = make([]byte, 0)
		return 0, nil, MaxPackError
	}
	if len(*data) < 6 {
		//log.Warnf("package is not complete")
		return 0, nil, nil
	}
	clen := int(binary.LittleEndian.Uint32((*data)[:4]))
	//log.Debugf("clen=%+v", clen)
	if len(*data) < clen + 4 {
		//log.Warnf("package is not complete")
		log.Errorf("data len error %v < %v : %+v", len(*data), clen + 4, *data)
		//*data = make([]byte, 0)
		return 0, nil, DataLenError
	}
	//log.Debugf("cmd=%+v", (*data)[4:6])
	cmd     := int(binary.LittleEndian.Uint16((*data)[4:6]))
	//log.Debugf("content=%+v === %v", (*data)[6 : clen + 4], string((*data)[6 : clen + 4]))
	content := make([]byte, len((*data)[6 : clen + 4]))
	//content := data[6 : clen + 4]
	copy(content, (*data)[6 : clen + 4])

	//if len(*data) < clen + 4 {
	//	log.Errorf("package error")
	//	*data  = append((*data)[:0], (*data)[len(*data):]...)
	//} else {
	//log.Debugf("data==%+v, %+v==%+v",clen+4, *data, string(*data))
	*data  = append((*data)[:0], (*data)[clen+4:]...)
	//}
	//tcp.buffer = append(tcp.buffer[:0], tcp.buffer[end:]...)
	//log.Debugf("return(%+v)(%+v)(%+v)", cmd, content, nil)
	return cmd, content, nil
}
