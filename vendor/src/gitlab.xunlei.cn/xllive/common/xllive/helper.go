package xllive

import (
	"fmt"
	"crypto/md5"
	"strconv"
	"strings"
	"errors"
)

func HashId(id int64, max int64) int64 {
	idstr   := []byte(fmt.Sprintf("%v", id))
	has     := md5.Sum(idstr)
	md5str  := fmt.Sprintf("%x", has)
	v1      := md5str[:2]
	v2      := md5str[len(md5str)-2:]
	i16, _  := strconv.ParseInt(v1+v2, 16, 64)
	return i16 % max + 1
}

func ParseRoomId(roomid string) (id int64, userid int64, err error) {
	temp := strings.Split(roomid, "_")
	if len(temp) != 2 {
		return 0, 0, errors.New("roomid["+roomid+"] error")
	}
	//var err error
	id, err = strconv.ParseInt(temp[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	userid, err = strconv.ParseInt(temp[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return id, userid, nil
}

