package db

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"
	"vtmmock/helper"
	. "vtmmock/log"

	//log "github.com/sirupsen/logrus"
	"github.com/subchen/go-xmldom"
)

type TemplateReq struct {
	Ty           string           `json:"ty"`
	TemplateName string           `json:"templateName"`
	Data         string           `json:"data"`
	Ch           chan TemplateRes `json:"resCh,omitempty"`
}

type Cmd struct {
	Request Req
	Res     chan Res
}

type Req struct {
	Ty   string `json:"ty"`
	Id   string `json:"id"`
	Data Rule   `json:"data"`
}

// 规则

type Rule struct {
	OrCondition []Condition `json:"orCondition"`
	Actions     []Action    `json:"actions"`
	ExpireSec   int         `json:"expiresec" default:"3600"`
	Id          string      `json:"id"`
}

type Condition struct {
	AndCondition []Item `json:"andCondition"`
}

type Item struct {
	Xpath    string `json:"xpath"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// Mode:
//       3，pre
//       4，forward
//       5，conclose
//       6，mock
//       7，post
//       8，timeout
type Action struct {
	Mode  string `json:"mode"`
	Items []Item `json:"items,omitempty"`
	//复用该字段作为超时时间
	TemplateName string `json:"templateName,omitempty"`
	Forward      string `json:"forward,omitempty"`
}

// response
type Res struct {
	Code  int    `json:"code"`
	Datas []Rule `json:"datas"`
}

type TemplateRes struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

// key: guid
// v: command
var records = make(map[string]Cmd)
var templates = make(map[string]TemplateReq)
var DbCh = make(chan Cmd)
var TemplatesCh = make(chan TemplateReq)
var cleanCycle = 3600

func SetCleanCycle(n int) {
	cleanCycle = n
}

func StartDb() {
	var res Res
	go cleaner()
	go StartTemplate()
	ruleCounter := 1

	for cmd := range DbCh {
		if cmd.Request.Ty == "query" {
			if v, ok := records[cmd.Request.Id]; ok {
				res.Code = 0
				res.Datas = []Rule{v.Request.Data}
				cmd.Res <- res
			} else if cmd.Request.Id == "" {
				res.Code = 1
				var datas []Rule
				for _, v := range records {
					datas = append(datas, v.Request.Data)
				}
				res.Datas = datas
				cmd.Res <- res
			} else {
				res.Code = -1
				res.Datas = []Rule{}
				cmd.Res <- res
			}
		} else if cmd.Request.Ty == "remove" {
			if _, ok := records[cmd.Request.Id]; ok {
				delete(records, cmd.Request.Id)
				res.Code = 0
				res.Datas = nil
				cmd.Res <- res
			} else {
				res.Code = -1
				res.Datas = nil
				cmd.Res <- res
			}
		} else if cmd.Request.Ty == "add" {

			// 过期时间设置默认值
			if cmd.Request.Data.ExpireSec == 0 {
				cmd.Request.Data.ExpireSec = cleanCycle
			}

			// 指定id default不生产递增id，会覆盖原default操作
			if cmd.Request.Data.Id != "default" {
				cmd.Request.Data.Id = strconv.Itoa(ruleCounter)
				ruleCounter++
			}

			records[cmd.Request.Data.Id] = cmd

			res.Code = 0
			res.Datas = []Rule{cmd.Request.Data}
			cmd.Res <- res
		} else if cmd.Request.Ty == "Expire" {
			Logger.Println("cleaner starting...")
			for k, v := range records {
				decreaseNum, _ := strconv.Atoi(cmd.Request.Id)
				v.Request.Data.ExpireSec -= decreaseNum
				if v.Request.Data.ExpireSec < 0 {
					delete(records, k)
					Logger.Println("cleaner removed key :" + string(k))
				}
			}
			Logger.Println("cleaner job done...")
			res.Code = 0
			res.Datas = []Rule{}
			cmd.Res <- res
		}

	}
}

func StartTemplate() {
	var res TemplateRes
	for cmd := range TemplatesCh {
		if cmd.Ty == "query" {
			if v, ok := templates[cmd.TemplateName]; ok {
				res.Code = 0
				res.Data = v.Data
				cmd.Ch <- res
			} else {
				res.Code = -1
				res.Data = ""
				cmd.Ch <- res
			}
		} else if cmd.Ty == "remove" {
			delete(templates, cmd.TemplateName)
			res.Code = 0
			res.Data = ""
			cmd.Ch <- res
		} else if cmd.Ty == "add" {
			if cmd.TemplateName == "" {
				res.Code = -1
				res.Data = "template name required"
				cmd.Ch <- res
			} else {
				templates[cmd.TemplateName] = cmd
				res.Code = 0
				res.Data = ""
				cmd.Ch <- res
			}
		}
	}
}

func uuid() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func Engine(frame []byte, uuid string) (response []byte, err error) {
	var tempResult []byte
	for _, record := range records {
		if match(record.Request.Data, frame, uuid) {
			Logger.Printf("[%s]match rule %s", uuid, record.Request.Data.Id)
			for _, v := range record.Request.Data.Actions {
				tempResult, err = doActions(v, frame, tempResult, uuid)
				//todo：细化错误处理
				if err != nil {
					return errorResponse(frame, err, uuid), err
				}
			}
			return tempResult, err
		}

	}
	//todo：没有任何规则匹配到的报文处理逻辑，目前只是报错，待添加默认动作
	return defaultAction(frame, tempResult, uuid)
}

func defaultAction(frame []byte, extFrame []byte, uuid string) (ret []byte, err error) {
	var tempResult []byte
	if v, ok := records["default"]; ok {
		Logger.Printf("[%s]%s", uuid, "default action performed.")
		for _, v := range v.Request.Data.Actions {
			tempResult, err = doActions(v, frame, extFrame, uuid)
			//todo：细化错误处理
			if err != nil {
				return errorResponse(frame, err, uuid), err
			}
		}
		return tempResult, err
	} else {
		return nil, fmt.Errorf("no default rule found, drop.")
	}
}

// todo: 一个action支持多个pre，post动作，目前一个action只有第一个pre，post有效
func doActions(action Action, frame []byte, extFrame []byte, uuid string) (ret []byte, err error) {
	Logger.Printf("[%s]%s", uuid, "do "+action.Mode)
	if action.Mode == "pre" {
		return xpathProcess(action.Items, frame, extFrame, uuid)
	} else if action.Mode == "post" {
		return xpathProcess(action.Items, extFrame, frame, uuid)
	} else if action.Mode == "mock" {
		return getTemplate(action.TemplateName, uuid)
	} else if action.Mode == "forward" {
		return getForwardResponse(action.Forward, frame, uuid)
	} else if action.Mode == "timeout" {
		t, _ := strconv.Atoi(action.TemplateName)
		time.Sleep(time.Duration(t) * time.Second)
		Logger.Printf("[%s]%s", uuid, "sleep(s):"+action.TemplateName)
		return extFrame, nil
	} else if action.Mode == "conClose" {
		return nil, fmt.Errorf("conClose")
	} else {
		return nil, fmt.Errorf("unknown mode [" + action.Mode + "] found")
	}
}

func getForwardResponse(s string, frame []byte, uuid string) (ret []byte, err error) {
	buf := make([]byte, 1024)
	conn, err := net.Dial("tcp", s)
	Logger.Printf("[%s]%s", uuid, "connect "+s)
	if err != nil {
		Logger.Printf("[%s]%s", uuid, err.Error())
		return nil, fmt.Errorf("connecting failed :" + s)
	}

	defer conn.Close()

	// forward the request
	Logger.Printf("[%s]%s", uuid, "forward : "+string(helper.GenFrame(frame)))
	conn.Write(helper.GenFrame(frame))
	n := 0
	prefixLen := 6
	var response []byte
	// 设置接收超时时间
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	for {
		// 确认前缀完整收到
		if n < prefixLen {
			conter, err := conn.Read(buf)
			if err != nil {
				return nil, err
			}

			n = n + conter

			response = append(response, buf[:conter]...)
			continue
		}

		frameLen, err := strconv.Atoi(string(helper.GetFrameLengthSlice(response)))

		if err != nil {
			return nil, err
		}

		if n < prefixLen+frameLen {
			conter, err := conn.Read(buf)
			if err != nil {
				return nil, err
			}

			n = n + conter

			response = append(response, buf[:conter]...)
			continue
		}
		Logger.Printf("[%s]%s", uuid, "forward response : "+string(response[:prefixLen+frameLen]))
		response = response[prefixLen : prefixLen+frameLen]
		break
	}

	return response, nil
}

func getTemplate(s, uuid string) (ret []byte, err error) {
	var req TemplateReq
	ch := make(chan TemplateRes)
	req.Ch = ch
	req.TemplateName = s
	req.Ty = "query"

	TemplatesCh <- req

	select {
	case res := <-ch:
		if res.Code != -1 {
			Logger.Printf("[%s]%s", uuid, s+" found in cache")
			return []byte(res.Data), nil
		}
	case <-time.After(3 * time.Second):
	}
	return getTemplateFromFile(s, uuid)
}

// todo:当前模板以本地文件形式存储，后续将存到缓存redis中
func getTemplateFromFile(s, uuid string) (ret []byte, err error) {
	content, err := ioutil.ReadFile(s)
	if err != nil {
		Logger.Printf("[%s]%s", uuid, s+" not found , use default template ")
		content, err = ioutil.ReadFile("error.xml")
	}
	return content, err
}

func xpathProcess(items []Item, data []byte, extData []byte, uuid string) (ret []byte, err error) {
	var document *xmldom.Node
	defer func() {
		if e := recover(); e != nil {
			ret = nil
			err = fmt.Errorf("%s", e)
			Logger.Printf("[%s]%s", uuid, err.Error())
		}
	}()
	if doc, err := xmldom.ParseXML(string(data)); doc == nil {
		return nil, fmt.Errorf("xml frame not well formed")
	} else {
		document = xmldom.Must(doc, err).Root
	}
	err = nil
	for _, v := range items {
		node := document.QueryOne(v.Xpath)

		if node == nil {
			return nil, fmt.Errorf(v.Xpath + " not found")
		}

		if v.Operator == "set" {
			node.Text = v.Value
		} else if v.Operator == "eq" {
			if node.Text != v.Value {
				return nil, fmt.Errorf("not match")
			}
		} else if v.Operator == "gt" {
			if node.Text != v.Value {
				return nil, fmt.Errorf("not match")
			}
		} else if v.Operator == "lt" {
			if node.Text != v.Value {
				return nil, fmt.Errorf("not match")
			}
		} else if v.Operator == "Correlate" {
			doc := xmldom.Must(xmldom.ParseXML(string(extData))).Root
			root := doc.QueryOne(v.Value)
			if root != nil {
				node.Text = root.Text
			} else {
				node.Text = "ERROR: " + v.Value + " not found in request"
			}
		}
	}
	return []byte(document.XML()), err
}

func errorResponse(frame []byte, e error, uuid string) []byte {
	var content []byte
	content, _ = getTemplate("error.xml", uuid)

	item1 := Item{
		Xpath:    "//RET_MSG[1]",
		Operator: "set",
		Value:    e.Error(),
	}

	item2 := Item{
		Xpath:    "//SEQ_NO",
		Operator: "Correlate",
		Value:    "//SEQ_NO",
	}
	result, _ := xpathProcess([]Item{item1, item2}, content, frame, uuid)
	return result
}

func match(condition Rule, frame []byte, uuid string) bool {
	result := false
	Logger.Printf("[%s]testing rule %s", uuid, condition.Id)
	for _, v := range condition.OrCondition {
		if _, e := xpathProcess(v.AndCondition, frame, nil, uuid); e == nil {
			return true
		}
	}
	return result
}

func isDefault(condition Rule, frame []byte) bool {
	result := false
	for _, v := range condition.OrCondition {
		for _, i := range v.AndCondition {
			if i.Operator == "default" {
				return true
			}
		}
	}
	return result
}

func cleaner() {
	for {
		time.Sleep(time.Duration(600) * time.Second)
		var c Cmd
		ch := make(chan Res)
		c.Request.Data = Rule{}
		c.Request.Ty = "Expire"
		c.Request.Id = "600"
		c.Res = ch

		DbCh <- c
		_ = <-ch
		close(ch)
	}
}
