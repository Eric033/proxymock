package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
	"vtmmock/db"
	"vtmmock/helper"
	. "vtmmock/log"
	"vtmmock/mock"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func TestForward(t *testing.T) {
	go db.StartDb()
	go StartCmd("8000")
	go mock.StartMock("8001", "utf-8")
	go echo()
	time.Sleep(time.Duration(2) * time.Second)

	// add rules
	requsetBody := `{
		"ty":"add",
		"id":"1234",
		"datas":[
			{"orCondition":[
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParseXML']","operator":"eq","value":"abc"},{"xpath":"//testcase[@time='0.004']","operator":"eq","value":"abc"}]},
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParse']","operator":"eq","value":"123"},{"xpath":"//testcase[@time='0.005']","operator":"eq","value":"123"}]}
			 ],
			 "actions":[
					{"mode":"pre","items":[{"xpath":"//testcase[@id='ExampleParseXML']","operator":"set","value":"elva"}]},
					{"mode":"forward","Forward":"127.0.0.1:8808"},
					{"mode":"post","items":[{"xpath":"//testcase[@id='ExampleParseXML']","operator":"set","value":"eric"}]}
			 ]

			}
		]
	}`

	resp, err := http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("add response : " + (string(body)))

	// query rules
	requsetBody = `{
		"ty":"query",
		"id":"1234"
	}
	`
	resp, err = http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("search response : " + (string(body)))
	if find := strings.Contains(string(body), "testcase"); find == false {
		t.Error("search failed")
	}

	buf := make([]byte, 1024)
	conn, err := net.Dial("tcp", "127.0.0.1:8001")
	if err != nil {
		t.Error("connect failed")
		return
	}

	//准备命令行标准输入
	request := `00000348<testsuites><testsuite tests="2" failures="0" time="0.009" name="github.com/subchen/go-xmldom"><properties><property name="go.version">go1.8.1</property></properties><testcase classname="go-xmldom" id="ExampleParseXML" time="0.004">abc</testcase><testcase classname="go-xmldom" id="ExampleParse" time="0.005">123</testcase></testsuite></testsuites>`

	conn.Write([]byte(request))
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	//time.Sleep(time.Duration(5) * time.Second)
	n, e := conn.Read(buf)
	if e != nil {
		t.Error(e.Error())
	}
	t.Error(fmt.Sprintf("%d", n))
	t.Error(string(buf[:n]))
	if !strings.Contains(string(buf[:n]), "00000349") {
		t.Error("prefix mismatch")
	}
}

func TestMockGBK(t *testing.T) {
	go db.StartDb()
	go StartCmd("8000")
	go mock.StartMock("8001", "gbk")
	go echo()
	time.Sleep(time.Duration(2) * time.Second)

	// add rules
	requsetBody := `{
		"ty":"add",
		"id":"1234",
		"datas":[
			{"orCondition":[
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParseXML']","operator":"eq","value":"abc"},{"xpath":"//testcase[@time='0.004']","operator":"eq","value":"abc"}]},
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParse']","operator":"eq","value":"123"},{"xpath":"//testcase[@time='0.005']","operator":"eq","value":"123"}]}
			 ],
			 "actions":[ 
					{"mode":"pre","items":[{"xpath":"//testcase[@id='ExampleParseXML']","operator":"set","value":"elva"}]},
					{"mode":"mock","TemplateName":"test1.xml"},
					{"mode":"post","items":[{"xpath":"//testcase[2]","operator":"set","value":"boobooke"}]}
			 ]
						
			}
		]
	}`

	resp, err := http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("add response : " + (string(body)))

	// query rules
	requsetBody = `{
		"ty":"query",
		"id":"1234"
	}
	`
	resp, err = http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("search response : " + (string(body)))
	if find := strings.Contains(string(body), "testcase"); find == false {
		t.Error("search failed")
	}

	// add template
	requsetBody = `{
		"ty":"add",
		"templateName":"test.xml",
		"data":"<testsuites><testsuite><testcase>positive</testcase><testcase>nagtive</testcase></testsuite></testsuites>"
	}`

	resp, err = http.Post("http://127.0.0.1:8000/createNewTemplate",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("add response : " + (string(body)))

	buf := make([]byte, 1024)
	conn, err := net.Dial("tcp", "127.0.0.1:8001")
	if err != nil {
		t.Error("connect failed")
		return
	}
	//准备命令行标准输入
	request := utf8ToGbk([]byte(`<testsuites><testsuite tests="2" failures="0" time="0.009" name="github.com/subchen/go-xmldom"><properties><property name="go.version">go1.8.1</property></properties><testcase classname="go-xmldom" id="ExampleParseXML" time="0.004">你好啊</testcase><testcase classname="go-xmldom" id="ExampleParse" time="0.005">123</testcase></testsuite></testsuites>`))
	prefix := fmt.Sprintf("%08d", len(request))
	conn.Write(append([]byte(prefix), request...))
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	//time.Sleep(time.Duration(5) * time.Second)
	n, e := conn.Read(buf)
	if e != nil {
		t.Error(e.Error())
	}
	t.Error(string(buf[:n]))
	t.Error(string(gbkToUtf8(buf[:n])))
	if !strings.Contains(string(buf[:n]), "00000326") {
		t.Error("prefix mismatch")
	}
}

func utf8ToGbk(b []byte) []byte {
	r := bytes.NewReader(b)

	decoder := transform.NewReader(r, simplifiedchinese.GBK.NewEncoder()) //GB18030

	content, _ := ioutil.ReadAll(decoder)

	return content
}

func gbkToUtf8(b []byte) []byte {
	tfr := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(tfr)
	if e != nil {
		return nil
	}
	return d
}

func TestMock(t *testing.T) {
	go db.StartDb()
	go StartCmd("8000")
	go mock.StartMock("8001", "utf-8")
	go echo()
	time.Sleep(time.Duration(2) * time.Second)
	helper.SetPrefixLen(8)
	// add rules
	requsetBody := `{
		"ty":"add",
		"id":"1234",
		"datas":[
			{"orCondition":[
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParseXML']","operator":"eq","value":"abc"},{"xpath":"//testcase[@time='0.004']","operator":"eq","value":"abc"}]},
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParse']","operator":"eq","value":"123"},{"xpath":"//testcase[@time='0.005']","operator":"eq","value":"123"}]}
			 ],
			 "actions":[ 
					{"mode":"pre","items":[{"xpath":"//testcase[@id='ExampleParseXML']","operator":"set","value":"elva"}]},
					{"mode":"mock","TemplateName":"test.xml"},
					{"mode":"post","items":[{"xpath":"//testcase[2]","operator":"set","value":"boobooke"}]}
			 ]
						
			}
		]
	}`

	resp, err := http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("add response : " + (string(body)))

	// query rules
	requsetBody = `{
		"ty":"query",
		"id":"1234"
	}
	`
	resp, err = http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("search response : " + (string(body)))
	if find := strings.Contains(string(body), "testcase"); find == false {
		t.Error("search failed")
	}

	// add template
	requsetBody = `{
		"ty":"add",
		"templateName":"test.xml",
		"data":"<testsuites><testsuite><testcase>positive</testcase><testcase>nagtive</testcase></testsuite></testsuites>"
	}`

	resp, err = http.Post("http://127.0.0.1:8000/createNewTemplate",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("add response : " + (string(body)))

	buf := make([]byte, 1024)
	conn, err := net.Dial("tcp", "127.0.0.1:8001")
	if err != nil {
		t.Error("connect failed")
		return
	}
	//准备命令行标准输入
	request := `00000348<testsuites><testsuite tests="2" failures="0" time="0.009" name="github.com/subchen/go-xmldom"><properties><property name="go.version">go1.8.1</property></properties><testcase classname="go-xmldom" id="ExampleParseXML" time="0.004">abc</testcase><testcase classname="go-xmldom" id="ExampleParse" time="0.005">123</testcase></testsuite></testsuites>`

	conn.Write([]byte(request))
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	//time.Sleep(time.Duration(5) * time.Second)
	n, e := conn.Read(buf)
	if e != nil {
		t.Error(e.Error())
	}
	t.Error(string(buf[:n]))
	if !strings.Contains(string(buf[:n]), "00000326") {
		t.Error("prefix mismatch")
	}
}

func TestForwardPreError(t *testing.T) {
	go db.StartDb()
	go StartCmd("8000")
	go mock.StartMock("8001", "gbk")
	go echo()
	helper.SetPrefixLen(8)
	time.Sleep(time.Duration(2) * time.Second)

	// add rules
	requsetBody := `{
		"ty":"add",
		"id":"1234",
		"datas":[
			{"orCondition":[
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParseXML']","operator":"eq","value":"abc"},{"xpath":"//testcase[@time='0.004']","operator":"eq","value":"abc"}]},
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParse']","operator":"eq","value":"123"},{"xpath":"//testcase[@time='0.005']","operator":"eq","value":"123"}]}
			 ],
			 "actions":[ 
					{"mode":"pre","items":[{"xpath":"//testcase[@id='ExampleParseXMLddd']","operator":"set","value":"elva"}]},
					{"mode":"mock","TemplateName":"test2.xml"},
					{"mode":"post","items":[{"xpath":"//testcase[2]","operator":"set","value":"boobooke"}]}
			 ]
						
			}
		]
	}`

	resp, err := http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("add response : " + (string(body)))

	// query rules
	requsetBody = `{
		"ty":"query",
		"id":"1234"
	}
	`
	resp, err = http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("search response : " + (string(body)))
	if find := strings.Contains(string(body), "testcase"); find == false {
		t.Error("search failed")
	}

	buf := make([]byte, 1024)
	conn, err := net.Dial("tcp", "127.0.0.1:8001")
	if err != nil {
		t.Error("connect failed")
		return
	}

	//准备命令行标准输入
	request := `00000348<testsuites><testsuite tests="2" failures="0" time="0.009" name="github.com/subchen/go-xmldom"><properties><property name="go.version">go1.8.1</property></properties><testcase classname="go-xmldom" id="ExampleParseXML" time="0.004">abc</testcase><testcase classname="go-xmldom" id="ExampleParse" time="0.005">123</testcase></testsuite></testsuites>`

	conn.Write([]byte(request))
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	//time.Sleep(time.Duration(5) * time.Second)
	n, e := conn.Read(buf)
	if e != nil {
		t.Error(e.Error())
	}
	t.Error(fmt.Sprintf("%d", n))
	t.Error(string(buf[:n]))
	if !strings.Contains(string(buf[:n]), "00000105") {
		t.Error("prefix mismatch")
	}
}

func TestNotMatch(t *testing.T) {
	go db.StartDb()
	go StartCmd("8000")
	go mock.StartMock("8001", "utf-8")
	go echo()
	time.Sleep(time.Duration(2) * time.Second)

	// add rules
	requsetBody := `{ 
		"ty":"add",
		"id":"1234",
		"datas":[
			{"orCondition":[
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParseXMLd']","operator":"eq","value":"abc"},{"xpath":"//testcase[@time='0.004']","operator":"eq","value":"abc"}]},
					 {"andCondition":[{"xpath":"//testcase[@id='ExampleParsed']","operator":"eq","value":"123"},{"xpath":"//testcase[@time='0.005']","operator":"eq","value":"123"}]}
			 ],
			 "actions":[ 
					{"mode":"pre","items":[{"xpath":"//testcase[@id='ExampleParseXMLddd']","operator":"set","value":"elva"}]},
					{"mode":"mock","TemplateName":"test.xml"},
					{"mode":"post","items":[{"xpath":"//testcase[2]","operator":"set","value":"boobooke"}]}
			 ]
						
			}
		]
	}`

	resp, err := http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("add response : " + (string(body)))

	// query rules
	requsetBody = `{
		"ty":"query",
		"id":"1234"
	}
	`
	resp, err = http.Post("http://127.0.0.1:8000/createNewRules",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("search response : " + (string(body)))
	if find := strings.Contains(string(body), "testcase"); find == false {
		t.Error("search failed")
	}

	buf := make([]byte, 1024)
	conn, err := net.Dial("tcp", "127.0.0.1:8001")
	if err != nil {
		t.Error("connect failed")
		return
	}

	//准备命令行标准输入
	request := `00000371<testsuites><testsuite tests="2" failures="0" time="0.009" name="github.com/subchen/go-xmldom"><properties><property name="go.version">go1.8.1</property></properties><testcase classname="go-xmldom" id="ExampleParseXML" time="0.004">abc</testcase><testcase classname="go-xmldom" id="ExampleParse" time="0.005">123</testcase><SEQ_NO>999999</SEQ_NO></testsuite></testsuites>`

	conn.Write([]byte(request))
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	//time.Sleep(time.Duration(10) * time.Second)
	n, e := conn.Read(buf)
	if e != nil {
		t.Error(e.Error())
	}

	Logger.Println(string(buf[:n]))
	if !strings.Contains(string(buf[:n]), "not matched") {
		t.Error("prefix mismatch")
	}
}

func echo() {
	listener, _ := net.Listen("tcp", "127.0.0.1:8808")

	for {
		conn, _ := listener.Accept()
		buf := make([]byte, 1024)

		conter, _ := conn.Read(buf)

		conn.Write(buf[:conter])

		conn.Close()
		Logger.Println("echo done :" + string(helper.GetFrameLengthSlice(buf)))
		Logger.Println("echo done :" + string(buf[:conter]))
	}
}

func TestTemplateAddQueryRemove(t *testing.T) {
	go db.StartTemplate()
	go StartCmd("8000")

	time.Sleep(time.Duration(2) * time.Second)

	// add rules
	requsetBody := `{
		"ty":"add",
		"templateName":"1234",
		"data":"<testsuites><testsuite><testcase>positive</testcase><testcase>nagtive</testcase></testsuite></testsuites>"
	}`

	resp, err := http.Post("http://127.0.0.1:8000/createNewTemplate",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("add response : " + (string(body)))

	// query rules
	requsetBody = `{
		"ty":"query",
		"templateName":"1234"
	}
	`
	resp, err = http.Post("http://127.0.0.1:8000/createNewTemplate",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	Logger.Println("query response : " + (string(body)))
	requsetBody = `{
		"ty":"remove",
		"templateName":"1234"
	}
	`
	resp, err = http.Post("http://127.0.0.1:8000/createNewTemplate",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	Logger.Println("remove response : " + (string(body)))

	requsetBody = `{
		"ty":"query",
		"templateName":"1234"
	}
	`
	resp, err = http.Post("http://127.0.0.1:8000/createNewTemplate",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("second search response : " + (string(body)))
}

func TestTemplateNotFound(t *testing.T) {
	go db.StartTemplate()
	go StartCmd("8000")

	time.Sleep(time.Duration(2) * time.Second)

	// add rules
	requsetBody := `{
		"ty":"add",
		"data":"abcdefghi"
	}`

	resp, err := http.Post("http://127.0.0.1:8000/createNewTemplate",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("add response : " + (string(body)))

	// query rules
	requsetBody = `{
		"ty":"query",
		"id":"1234"
	}
	`
	resp, err = http.Post("http://127.0.0.1:8000/createNewTemplate",
		"application/json",
		strings.NewReader(requsetBody))
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	Logger.Println("search response : " + (string(body)))
}

func TestRex(t *testing.T) {
	str := "Welcome for Beijing-Tianjin CRH train."
	reg := regexp.MustCompile(" ")
	result := reg.ReplaceAllString(str, "@")
	t.Error(result) //将空格替换为@字符
}
