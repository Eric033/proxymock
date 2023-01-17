package mock

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"vtmmock/db"
	"vtmmock/helper"
	. "vtmmock/log"
)

func StartMock(port, encoding string) {
	listener, err := net.Listen("tcp", ":"+port)
	ServerHandleError(err, "net.listen")

	for {
		conn, e := listener.Accept()
		ServerHandleError(e, "listener.accept")
		go mock(conn, encoding)

	}
}

func mock(conn net.Conn, encoding string) {
	defer conn.Close() // 关闭conn
	n := 0
	prefixLen := 6

	// 创建一个新的切片
	var frame []byte
	buf := make([]byte, 2048)

	for {
		// 确认前缀完整收到
		if n < prefixLen {
			conter, err := conn.Read(buf)
			if err != nil {
				return
			}

			n = n + conter
			frame = append(frame, buf[:conter]...)
			continue
		}
		frameLen, err := strconv.Atoi(string(helper.GetFrameLengthSlice(frame)))

		if err != nil {
			return
		}

		// 确认报文收到完整
		if n < prefixLen+frameLen {
			conter, err := conn.Read(buf)
			if err != nil {
				Logger.Println(err.Error())
				return
			}

			n = n + conter
			frame = append(frame, buf[:conter]...)
			continue
		}

		break
	}
	// 处理报文内容
	uuid := helper.UUID()
	Logger.Printf("[%s]%s", uuid, "received <=== : "+string(frame[prefixLen:]))
	process(conn, frame[prefixLen:], encoding, uuid)
}

func processMock(conn net.Conn, frame []byte) {
	defer conn.Close()
	conn.Write(frame)
}

func process(conn net.Conn, frame []byte, encoding string, uuid string) {
	// mock engine
	var err error
	if strings.ToUpper(encoding) == "GBK" {
		frame = helper.GbkToUtf8(frame)
	}

	response, err := db.Engine(frame, uuid)

	if err != nil {
		Logger.Printf("[%s]%s", uuid, err.Error())
		conn.Close()
		return
	}

	if strings.ToUpper(encoding) == "GBK" {
		response = helper.Utf8ToGbk(response)
	}

	response = helper.GenFrame(response)
	if _, e := conn.Write(response); e != nil {
		Logger.Printf("[%s]%s", uuid, "reply failed: "+e.Error())
	} else {
		Logger.Printf("[%s]%s", uuid, "reply ===> : "+string(response))
	}

}

func ServerHandleError(err error, when string) {
	if err != nil {
		fmt.Println(err, when)
		os.Exit(1)
	}
}
