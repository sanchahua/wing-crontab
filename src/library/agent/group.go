package agent

import 	log "github.com/sirupsen/logrus"


func (c *TcpClients) append(node *tcpClientNode) {
	*c = append(*c, node)
}

func (c *TcpClients) send(data []byte) {
	for _, node := range *c {
		node.send(data)
	}
}

func (c *TcpClients) asyncSend(data []byte) {
	for _, node := range *c {
		//log.Debugf("%v node keepalive", key)
		node.AsyncSend(data)
	}
}

func (c *TcpClients) remove(node *tcpClientNode) {
	for index, n := range *c {
		if n == node {
			*c = append((*c)[:index], (*c)[index+1:]...)
			break
		}
	}
	log.Debugf("#####################remove node, current len %v", len(*c))
}

func (c *TcpClients) close() {
	for _, node := range *c {
		node.close()
	}
}