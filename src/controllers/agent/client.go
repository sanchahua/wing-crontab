package agent

import (
	"time"
	log "github.com/sirupsen/logrus"
	"models/cron"
	"github.com/jilieryuyi/wing-go/tcp"
)

func (c *Controller) onClientEvent(tcp *tcp.Client, content []byte) {
	log.Infof("==========onClientEvent==========")
	cmd, data, err := c.codec.Decode(content)
	if err != nil {
		log.Errorf("%v, %+v, %v", err, content, string(content))
		return
	}
	log.Infof("cmd=%v,data=%+v", cmd, data)

	sendData, _ := decodeSendData(data)

	switch cmd {
	case CMD_RUN_COMMAND:
		log.Infof("======run command======")
		item, _:= decodeRunItem(sendData.Data)

		log.Infof("####################client receive command: %+v", sendData)
		log.Infof("####################client receive command: %+v", item)

		isMutex := byte(0)
		if item.IsMutex {
			isMutex = byte(1)
		}
		c.onCommand(item.Id, item.Command, sendData.Address, c.ctx.Config.BindAddress, isMutex, func() {
			tcp.Send(data)
		})
	}
}

// send data to leader
func (c *Controller) SyncToLeader(event int, row *cron.CronEntity) {
	// client发送到server，实际上这里的msgId没有用
	// client发送到server的时候会自动生成msgId
	r     := rowData{Event:event, Row:row,}
	rd, _ := r.encode()
	d     := newSendData(1, CMD_CRONTAB_CHANGE, rd, nil, 0, false, c.ctx.Config.BindAddress)
	jd    := d.encode()

	sendData, _      := c.codec.Encode(d.Cmd, jd)
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
	sd, _ := c.codec.Encode(CMD_PULL_COMMAND, []byte(""))
	c.client.AsyncSend(sd)
}

