package db

import (
	"fmt"
	"strings"
	"testing"
	_ "vtmmock/log"
)

func TestQuery_nodata_in_db(t *testing.T) {
	var c Cmd
	ch := make(chan Res)
	c.Request.Data = Rule{}
	c.Request.Ty = "query"
	c.Request.Id = ""
	c.Res = ch

	go StartDb()

	DbCh <- c

	res := <-ch

	if res.Code != -1 || len(res.Datas) != 0 {
		t.Errorf("%v", res)
	}
}

func TestQuery_found_in_db(t *testing.T) {
	var c Cmd
	ch := make(chan Res)
	c.Request.Data = Rule{}
	c.Request.Ty = "query"
	c.Request.Id = "1234"
	c.Res = ch

	go StartDb()

	DbCh <- c
	t.Error(fmt.Sprintf("%d", 7))
	res := <-ch

	if res.Code != -1 || len(res.Datas) != 0 {
		t.Errorf("%v", res)
	}
}

func TestXmldom(t *testing.T) {
	xml := `
	<testsuites>
		<aaa>
			<testcase classname="go-xmldom" id="ExampleParseXML" time="0.004"></testcase>
			<testcase classname="go-xmldom" id="ExampleParse" time="0.005"></testcase>
		</aaa>
		<SEQ_NO>1234567890</SEQ_NO>
	</testsuites>`
	//xml := `<testsuites><testcase classname="go-xmldom" id="ExampleParseXML" time="0.004">abc</testcase><testcase classname="go-xmldom" id="ExampleParse" time="0.005">123</testcase></testsuites>`

	item1 := Item{
		Xpath:    "//testcase[@id='ExampleParseXML']",
		Operator: "set",
		Value:    "123",
	}

	item2 := Item{
		Xpath:    "//testcase[2]",
		Operator: "set",
		Value:    "abcd",
	}
	result, _ := xpathProcess([]Item{item1, item2}, []byte(xml), nil, "")
	if !strings.Contains(string(result), "abcd") {
		t.Error(string(result))
	}

}

func TestForward(t *testing.T) {
	// response, _ := getForwardResponse("127.0.0.1:80", []byte{'a', 'b'})
	// t.Error(string(response))
	n := 340
	d := fmt.Sprintf("%08d", n)
	result := append([]byte(d), []byte("abcd")...)
	t.Error(string(result))
}

func TestGegtemplate(t *testing.T) {
	d, _ := getTemplate("test.xml", "")
	if !strings.Contains(string(d), "positive") {
		t.Error(string(d))
	}
}

func TestErrorResponse(t *testing.T) {
	xml := `
	<testsuites>
		<aaa>
			<testcase classname="go-xmldom" id="ExampleParseXML" time="0.004"></testcase>
			<testcase classname="go-xmldom" id="ExampleParse" time="0.005"></testcase>
		</aaa>
		<SEQ_NO>1234567890</SEQ_NO>
	</testsuites>`
	data := errorResponse([]byte(xml), fmt.Errorf("unittest"), "")
	if !strings.Contains(string(data), "unittest") || !strings.Contains(string(data), "1234567890") {
		t.Error(string(data))
	}
}

func TestTemp(t *testing.T) {
	records := map[int]string{}
	records = nil
	for _, v := range records {
		t.Error(string(v))
	}
}
