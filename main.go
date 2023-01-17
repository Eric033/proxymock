package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"vtmmock/cmd"
	"vtmmock/db"
	"vtmmock/helper"
	. "vtmmock/log"
	"vtmmock/mock"
)

func main() {
	var (
		cmdPort   string
		port      string
		encoding  string
		prefixLen int
		testMode  bool
		expireSec int
	)

	c := make(chan os.Signal)
	//signal.Notify(c)
	signal.Notify(c, syscall.SIGKILL)
	flag.StringVar(&cmdPort, "cp", "9091", "command port(http default 9091)")
	flag.StringVar(&port, "mp", "9090", "mock port(tcp default 9090)")
	flag.StringVar(&encoding, "enc", "utf-8", "mock encoding:utf-8 , gbk")
	flag.IntVar(&prefixLen, "pre", 6, "frame prefix length")
	flag.IntVar(&expireSec, "exp", 3600, "rule default expire time in second")
	flag.BoolVar(&testMode, "test", true, "true : start with a echo sever for testing")
	flag.Parse()
	Logger.Println("vtm mock starting ...")
	Logger.Println("cmdport: " + cmdPort)
	Logger.Println("port: " + port)
	Logger.Println("encoding: " + encoding)
	Logger.Printf("prefixLen: %d", prefixLen)

	if testMode {
		Logger.Println("testmode")
		go startEchoServer()
	}
	db.SetCleanCycle(expireSec)
	// 待参数化
	helper.SetPrefixLen(prefixLen)

	go cmd.StartCmd(cmdPort)
	go mock.StartMock(port, encoding)
	go db.StartDb()

	// don't exit
	s := <-c

	panic(s)
}

// just for test
func startEchoServer() {
	listener, _ := net.Listen("tcp", ":8808")
	Logger.Println("echo server 8808")
	for {
		conn, _ := listener.Accept()
		buf := make([]byte, 2048)
		conter, _ := conn.Read(buf)
		conn.Write(buf[:conter])
		conn.Close()

		Logger.Println("echo done :" + string(buf[:conter]))
	}
}
