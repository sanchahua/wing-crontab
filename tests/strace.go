package main

import (
	"bufio"
	"strings"
	"bytes"
	"fmt"
	"flag"
	"os/exec"
	"os"
)

type Server struct {
	Ip string
	Port string
	Sorce string
}
func GetServers(straceOutput string) []*Server {
	r 	:= strings.NewReader(straceOutput)
	bio := bufio.NewReader(r)
	res := make([]*Server, 0)

	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			break
		}
		if bytes.HasPrefix(line, []byte("connect")) && bytes.HasSuffix(line, []byte("(Operation now in progress)")) {
			startIndex := bytes.Index([]byte(straceOutput), line)+len(line)+1
			st := bytes.Index(line, []byte("connect")) + len([]byte("connect"))+1
			socketSource := string(line[st:st+1])
			st = bytes.Index(line, []byte("sin_port=htons(")) + len([]byte("sin_port=htons("))
			port := line[st:][0:bytes.Index(line[st:], []byte(")"))]
			st = bytes.Index(line, []byte("sin_addr=inet_addr(\"")) + len([]byte("sin_addr=inet_addr(\""))
			ip := line[st:][0:bytes.Index(line[st:], []byte(")"))-1]
			straceOutput = straceOutput[startIndex:]
			bio.Reset(strings.NewReader(straceOutput))
			res = append(res, &Server{
				Ip: string(ip),
				Port: string(port),
				Sorce: string(socketSource),
			})
		}
	}
	return res
}

func GetReadWriteTimes(straceOutput, source string) int {
	r 	     := strings.NewReader(straceOutput)
	bio      := bufio.NewReader(r)
	sendTime := 0

	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			break
		}
		if strings.Index(string(line), "sendto("+string(source)) == 0 {
			sendTime++
		}
	}
	return sendTime
}

func main() {
	var command string
	flag.StringVar(&command,"cmd", "", "")
	flag.Parse()
	var cmd *exec.Cmd
	var err error
	var straceOutputr []byte
	if command != "" {
		cmd = exec.Command("bash", "-c", "strace -v -f -e trace=network "+command)
		straceOutputr, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	straceOutput := string(straceOutputr)

	if straceOutput == "" {
		straceOutput = `socket(PF_INET6, SOCK_DGRAM, IPPROTO_IP) = 3
socket(PF_LOCAL, SOCK_STREAM|SOCK_CLOEXEC|SOCK_NONBLOCK, 0) = 3
connect(3, {sa_family=AF_LOCAL, sun_path="/var/run/nscd/socket"}, 110) = -1 ENOENT (No such file or directory)
socket(PF_LOCAL, SOCK_STREAM|SOCK_CLOEXEC|SOCK_NONBLOCK, 0) = 3
connect(3, {sa_family=AF_LOCAL, sun_path="/var/run/nscd/socket"}, 110) = -1 ENOENT (No such file or directory)
socket(PF_INET, SOCK_STREAM, IPPROTO_IP) = 3
connect(3, {sa_family=AF_INET, sin_port=htons(3306), sin_addr=inet_addr("10.10.62.28")}, 16) = -1 EINPROGRESS (Operation now in progress)
getsockopt(3, SOL_SOCKET, SO_ERROR, [0], [4]) = 0
setsockopt(3, SOL_TCP, TCP_NODELAY, [1], 4) = 0
recvfrom(3, "N\0\0\0", 4, MSG_DONTWAIT, NULL, NULL) = 4
recvfrom(3, "\n5.6.25-log\0>\4g\5?Y#?Ti3a\0\377\367-\2\0\177\200"..., 82, MSG_DONTWAIT, NULL, NULL) = 78
sendto(3, "n\0\0\1\215\242\v\0\0\0\0\300!\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0"..., 114, MSG_DONTWAIT, NULL, 0) = 114
recvfrom(3, "\7\0\0\2", 4, MSG_DONTWAIT, NULL, NULL) = 4
recvfrom(3, "\0\0\0\2\0\0\0", 82, MSG_DONTWAIT, NULL, NULL) = 7
sendto(3, "\17\0\0\0\3SET NAMES utf8", 19, MSG_DONTWAIT, NULL, 0) = 19
recvfrom(3, "\7\0\0\1\0\0\0\2\0\0\0", 75, MSG_DONTWAIT, NULL, NULL) = 11
socket(PF_INET, SOCK_STREAM, IPPROTO_IP) = 4
connect(4, {sa_family=AF_INET, sin_port=htons(8850), sin_addr=inet_addr("10.10.62.31")}, 16) = -1 EINPROGRESS (Operation now in progress)
getsockopt(4, SOL_SOCKET, SO_ERROR, [0], [4]) = 0
setsockopt(4, SOL_TCP, TCP_NODELAY, [1], 4) = 0
sendto(4, "4\nhget\n24\nshowapp_player_info_ha"..., 47, MSG_DONTWAIT, NULL, 0) = 47
recvfrom(4, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "ok\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(4, "9", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "9", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "8", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "{\"userid\":\"60000087\",\"total_gold"..., 999, MSG_DONTWAIT, NULL, NULL) = 999
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
socket(PF_INET, SOCK_STREAM, IPPROTO_IP) = 5
connect(5, {sa_family=AF_INET, sin_port=htons(8870), sin_addr=inet_addr("10.10.62.31")}, 16) = -1 EINPROGRESS (Operation now in progress)
getsockopt(5, SOL_SOCKET, SO_ERROR, [0], [4]) = 0
setsockopt(5, SOL_TCP, TCP_NODELAY, [1], 4) = 0
sendto(5, "10\nmulti_hget\n21\nshowapp_player_"..., 72, MSG_DONTWAIT, NULL, 0) = 72
recvfrom(5, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "ok\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(5, "1", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "60000087_0_1\n", 13, MSG_DONTWAIT, NULL, NULL) = 13
recvfrom(5, "6", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "410507\n", 7, MSG_DONTWAIT, NULL, NULL) = 7
recvfrom(5, "1", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "60000087_0_2\n", 13, MSG_DONTWAIT, NULL, NULL) = 13
recvfrom(5, "1", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "0\n", 2, MSG_DONTWAIT, NULL, NULL) = 2
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
sendto(5, "10\nmulti_hget\n21\nshowapp_player_"..., 72, MSG_DONTWAIT, NULL, 0) = 72
recvfrom(5, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "ok\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(5, "1", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "60000087_0_1\n", 13, MSG_DONTWAIT, NULL, NULL) = 13
recvfrom(5, "6", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "410507\n", 7, MSG_DONTWAIT, NULL, NULL) = 7
recvfrom(5, "1", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "60000087_0_2\n", 13, MSG_DONTWAIT, NULL, NULL) = 13
recvfrom(5, "1", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "0\n", 2, MSG_DONTWAIT, NULL, NULL) = 2
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
sendto(4, "4\nhget\n29\nshowapp_player_last7da"..., 52, MSG_DONTWAIT, NULL, 0) = 52
recvfrom(4, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "ok\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(4, "8", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "\"307760\"\n", 9, MSG_DONTWAIT, NULL, NULL) = 9
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
sendto(5, "10\nmulti_hget\n25\nshowapp_player_"..., 76, MSG_DONTWAIT, NULL, 0) = 76
recvfrom(5, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "ok\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(5, "1", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "60000087_0_2\n", 13, MSG_DONTWAIT, NULL, NULL) = 13
recvfrom(5, "1", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "0\n", 2, MSG_DONTWAIT, NULL, NULL) = 2
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
sendto(4, "4\nhget\n24\nshowapp_player_info_ha"..., 47, MSG_DONTWAIT, NULL, 0) = 47
recvfrom(4, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "ok\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(4, "9", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "9", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "8", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "{\"userid\":\"60000087\",\"total_gold"..., 999, MSG_DONTWAIT, NULL, NULL) = 999
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
socket(PF_INET, SOCK_STREAM, IPPROTO_IP) = 6
connect(6, {sa_family=AF_INET, sin_port=htons(8860), sin_addr=inet_addr("10.10.62.31")}, 16) = -1 EINPROGRESS (Operation now in progress)
getsockopt(6, SOL_SOCKET, SO_ERROR, [0], [4]) = 0
setsockopt(6, SOL_TCP, TCP_NODELAY, [1], 4) = 0
sendto(6, "4\nhget\n17\nlive_player_level\n8\n60"..., 40, MSG_DONTWAIT, NULL, 0) = 40
recvfrom(6, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(6, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(6, "ok\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(6, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(6, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(6, "12\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(6, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
sendto(4, "3\nget\n35\nshowapp_hot_player_key_"..., 46, MSG_DONTWAIT, NULL, 0) = 46
recvfrom(4, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "ok\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(4, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(4, "[]\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(4, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
sendto(5, "3\nget\n27\nshowapp_user_title_6000"..., 38, MSG_DONTWAIT, NULL, 0) = 38
recvfrom(5, "9", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "not_found\n", 10, MSG_DONTWAIT, NULL, NULL) = 10
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
sendto(3, "\230\0\0\0\3 select userid,type,sub_i"..., 156, MSG_DONTWAIT, NULL, 0) = 156
recvfrom(3, "\1\0\0\1\5C\0\0\2\3def\7showapp\ruser_title"..., 64, MSG_DONTWAIT, NULL, NULL) = 64
recvfrom(3, "?\0\v\0\0\0\3\t@\0\0\0?\0\0\3\3def\7showapp\ruse"..., 82, MSG_DONTWAIT, NULL, NULL) = 82
recvfrom(3, "\4\3def\7showapp\ruser_title_v2\ruser"..., 82, MSG_DONTWAIT, NULL, NULL) = 82
recvfrom(3, "pp\ruser_title_v2\ruser_title_v2\5v"..., 82, MSG_DONTWAIT, NULL, NULL) = 82
recvfrom(3, "_v2\ruser_title_v2\10selected\10selec"..., 82, MSG_DONTWAIT, NULL, NULL) = 66
sendto(5, "4\nsetx\n27\nshowapp_user_title_600"..., 49, MSG_DONTWAIT, NULL, 0) = 49
recvfrom(5, "2", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "ok\n", 3, MSG_DONTWAIT, NULL, NULL) = 3
recvfrom(5, "1", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
recvfrom(5, "1\n", 2, MSG_DONTWAIT, NULL, NULL) = 2
recvfrom(5, "\n", 1, MSG_DONTWAIT, NULL, NULL) = 1
`
	}

	servers := GetServers(straceOutput)
	for _, sev := range servers  {
		fmt.Fprintf(os.Stderr, "%v:%v, %v, read write times: %v\r\n", sev.Ip, sev.Port, sev.Sorce, GetReadWriteTimes(straceOutput, sev.Sorce))
	}
}
