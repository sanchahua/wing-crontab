package agent

import (
	"time"
	log "github.com/sirupsen/logrus"
	"models/cron"
	"github.com/jilieryuyi/wing-go/tcp"
)

func (c *Controller) onClientEvent(tcp *tcp.Client, content []byte) {
	cmd, data, err := c.codec.Decode(content)
	if err != nil {
		log.Errorf("%v", err)
		return
	}

	switch cmd {
	case CMD_RUN_COMMAND:
		//err := json.Unmarshal(content, &sendData)
		//if err != nil {
		//	log.Errorf("json.Unmarshal with %v", err)
		//	return
		//}
		//id, isMutex, command, dispatchServer, err := unpack(sendData.Data)
		//if err != nil {
		//	log.Errorf("%v", err)
		//	return
		//}
		//fmt.Fprintf(os.Stderr, "receive command, %v, %v, %v, %v, %v\r\n", id, isMutex, command, dispatchServer, err)
		//sdata := make([]byte, 0)
		//sid   := make([]byte, 8)
		//binary.LittleEndian.PutUint64(sid, uint64(id))
		//sdata = append(sdata, sid...)
		//sdata = append(sdata, isMutex)
		//sdata = append(sdata, []byte(sendData.Unique)...)
		sendData := data.(*SendData) //var sendData SendData
		item := sendData.Data.(*runItem)
		isMutex := byte(0)
		if item.isMutex {
			isMutex = byte(1)
		}
		c.onCommand(item.id, item.command, sendData.Address, c.ctx.Config.BindAddress, isMutex, func() {
			sd, _ := c.codec.Encode(CMD_RUN_COMMAND, sendData)
			tcp.Send(sd)
		})
		//fmt.Fprintf(os.Stderr, "receive command run end, %v, %v, %v, %v, %v\r\n", id, isMutex, command, dispatchServer, err)
		//case CMD_CRONTAB_CHANGE_OK:
		//	log.Infof("cron send to leader server ok (will delete from send queue): %+v", string(content))
		//	c.delSendQueueChan <- string(content)
		//case CMD_CRONTAB_CHANGE:
		//	//var sdata SendData
		//	//err := json.Unmarshal(content, &sdata)
		//	//if err != nil {
		//	//	log.Errorf("%+v", err)
		//	//} else {
		//		event := binary.LittleEndian.Uint32(sdata.Data[:4])
		//		go c.onCronChange(int(event), sdata.Data[4:])
		//	//}
		//}
	}
}

// send data to leader
func (c *Controller) SyncToLeader(event int, row *cron.CronEntity) {
	// client发送到server，实际上这里的msgId没有用
	// client发送到server的时候会自动生成msgId
	d := newSendData(1, CMD_CRONTAB_CHANGE, rowData{event:event, row:row,}, nil, 0, false, c.ctx.Config.BindAddress)
	sendData, _      := c.codec.Encode(d.Cmd, d)
	resource, _, err := c.client.Send(sendData)

	if err != nil {
		log.Error("SyncToLeader failure")
		return
	}

	// 这里采用同步发送，等待server端响应，响应超时时间设定为3秒
	res, err := resource.Wait(time.Second * 3)
	if err != nil {
		log.Error("SyncToLeader failure")
		return
	}
	log.Infof("SyncToLeader return: %v, %v", res, string(res))
}

// 这个api用来发送获取需要执行的定时任务
// 由crontab调用
// 一旦crontab执行完一定程度的定时任务，变得空闲就会主动获取新的定时任务
// 这个api就是发起主动获取请求
// 由client端发起
func (c *Controller) Pull() {
	sd, _ := c.codec.Encode(CMD_PULL_COMMAND, "")
	c.client.AsyncSend(sd)
}

