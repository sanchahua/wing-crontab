package agent

import (
	"encoding/binary"
	"time"
	"github.com/jilieryuyi/wing-go/tcp"
)

func (c *Controller) OnServerMessage(node *tcp.ClientNode, msgId int64, content []byte) {
	// content 二次解析后得到event
	// 这里的content全部使用json格式发送
	cmd, data, err := c.codec.Decode(content)
	if err != nil {
		return
	}
	switch cmd {
	case CMD_PULL_COMMAND:
		// server端收到pull请求
		// 这里的data是一个空字符串
		if len(c.onPullChan) < 32 {
			c.onPullChan <- message{node, msgId}
		}
	case CMD_CRONTAB_CHANGE:
		sendData, _ := decodeSendData(data)
		// 响应给客户端的请求
		// CMD_CRONTAB_CHANGE_OK客户端同步处理
		sd, err := c.codec.Encode(CMD_CRONTAB_CHANGE_OK, []byte(sendData.Unique))
		if err == nil {
			node.AsyncSend(msgId, sd)
		}
		// todo 如有必要，这里可以加一个广播，这样所有的节点都会收到定时任务改变事件
		// 触发定时任务改变事件
		row, _ := decodeRowData(sendData.Data)
		go c.onCronChange(row.Event, row.Row)
	case CMD_RUN_COMMAND:
		sendData, _ := decodeSendData(data)
		item, _:= decodeRunItem(sendData.Data)
		if item.IsMutex {
			sdata := make([]byte, 16)
			binary.LittleEndian.PutUint64(sdata[:8], uint64(item.Id))
			binary.LittleEndian.PutUint64(sdata[8:], uint64(int64(time.Now().UnixNano() / 1000000)))
			c.statisticsEndChan <- sdata
			c.runningEndChan <- item.Id
		}
		//c.delSendQueueChan <- unique
		//定时任务运行完返回server端（leader）
		c.OnCommandBack(item.Id, c.ctx.Config.BindAddress)
	}
}

