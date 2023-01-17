package mock

import (
	"net"
	"strings"
	"testing"
)

func Test_temp(t *testing.T) {
	go StartMock("8888", "utf-8")
	buf := make([]byte, 1024)
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		t.Error("connect failed")
		return
	}

	//准备命令行标准输入
	request := "00001063abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions00000003abcconditionsconditionsconditionsconditions"

	conn.Write([]byte(request))
	conn.Read(buf)
	t.Error(string(buf))
	if !strings.Contains(string(buf), "00001063") {
		t.Error("prefix mismatch")
	}

}
