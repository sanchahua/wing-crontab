package agent

import (
	"encoding/binary"
	"time"
	log "github.com/sirupsen/logrus"
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
		log.Info("server receive pull command")
		if len(c.onPullChan) < 32 {
			c.onPullChan <- message{node, msgId}
		}
	case CMD_CRONTAB_CHANGE:
		sdata, _ := decodeSendData(data)//data.(SendData)
		// 响应给客户端的请求
		// CMD_CRONTAB_CHANGE_OK客户端同步处理
		sd, err := c.codec.Encode(CMD_CRONTAB_CHANGE_OK, []byte(sdata.Unique))
		if err == nil {
			node.AsyncSend(msgId, sd)
		}
		// todo 如有必要，这里可以加一个广播，这样所有的节点都会收到定时任务改变事件
		// 触发定时任务改变事件
		row, _ := decodeRowData(sdata.Data)
		go c.onCronChange(row.Event, row.Row)
		//}
	case CMD_RUN_COMMAND:
		log.Infof("=====================================server receive run command")
		sendData, _ := decodeSendData(data)
		item, _:= decodeRunItem(sendData.Data)
		log.Infof("#######server run command %+v", *sendData)
		log.Infof("#######server run command %+v", *item)

		//id      := int64(binary.LittleEndian.Uint64(content[:8]))
		//isMutex := content[8]
		//unique  := string(content[9:])
		//fmt.Fprintf(os.Stderr, "receive run command end %v, %v, %v\r\n", id, isMutex, unique)

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

