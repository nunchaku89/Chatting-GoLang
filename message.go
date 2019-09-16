package main

import (
	"log"
	"net/http"
	"strings"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

var (
	clientMap    map[string]string
	clientString string
	server       *gosocketio.Server
)

type Message struct {
	ClientList string `json:"clientList"`
	Message    string `json:"message"`
	MyName     string `json:"myName"`
}

func main() {

	clientMap = make(map[string]string)
	server = gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	// 기본 이벤트
	server.On(gosocketio.OnConnection, connect)
	server.On(gosocketio.OnDisconnection, disConnect)
	server.On(gosocketio.OnError, socketError)

	// 커스텀 이벤트
	server.On("send", sendMsg)
	server.On("nick", setName)

	// 서버
	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)
	log.Panic(http.ListenAndServe("127.0.0.1:80", serveMux))
}

// 연결
func connect(c *gosocketio.Channel) {
	log.Println("Connected : ", c.Id())
	c.Join("chat")

	// Default 닉네임으로 채널 ID 사용
	clientMap[c.Id()] = c.Id()
	str := Message{
		ClientList: getlist(c),
		Message:    "<p>#<font color='blue'>" + clientMap[c.Id()] + "</font> 입장#</p>",
		MyName:     c.Id(),
	}
	channel, _ := server.GetChannel(c.Id())
	channel.Emit("changeMyName", str)
	c.BroadcastTo("chat", "connect", str)
}

// 연결 해제
func disConnect(c *gosocketio.Channel) {
	leaveName := clientMap[c.Id()]
	// 채팅방 나가기 & 사용자Map 에서 삭제
	c.Leave("chat")
	delete(clientMap, c.Id())
	str := Message{
		ClientList: getlist(c),
		Message:    "<p>#<font color='red'>" + leaveName + "</font> 퇴장#</p>",
	}

	c.BroadcastTo("chat", "message", str)
	log.Println("Disconnected : ", c.Id())
}

// 메시지 전송
func sendMsg(c *gosocketio.Channel, msg Message) string {
	msg.ClientList = getlist(c)
	msg.Message = "<p class='msg'>" + clientMap[c.Id()] + " : <font color='green'>" + msg.Message + "</font></p>"
	c.BroadcastTo("chat", "message", msg)
	return "OK"
}

// 닉네임 변경
func setName(c *gosocketio.Channel, nickname string) string {
	nickname = strings.TrimSpace(nickname)
	// 이름 중복체크
	for _, val := range clientMap {
		if val == nickname {
			return "X"
		}
	}

	// 사용자Map에 저장된 닉네임 변경
	originNickname := clientMap[c.Id()]
	clientMap[c.Id()] = nickname
	str := Message{
		ClientList: getlist(c),
		Message:    "<p>#<font color='red'> " + originNickname + "</font>이(가) <font color='blue'>" + clientMap[c.Id()] + "</font>로 닉네임 변경#</p>",
		MyName:     clientMap[c.Id()],
	}

	// 이름 바꾼 사용자 에게만 전송
	channel, _ := server.GetChannel(c.Id())
	channel.Emit("changeMyName", str)

	// 모든 방 Client 목록 갱신
	c.BroadcastTo("chat", "refreshClient", str)
	return "OK"
}

// 에러
func socketError(c *gosocketio.Channel) {
	log.Println("Error")
}

func getlist(c *gosocketio.Channel) string {
	str := ""
	for _, val := range clientMap {
		str += "<p class='clients'>" + val + "</p>"
		// if clientMap[c.Id()] != val {
		// 	str += "<p class='clients'>" + val + "</p>"
		// }

	}
	return str
}
