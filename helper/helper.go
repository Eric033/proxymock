package helper

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	. "vtmmock/log"

	"github.com/subchen/go-xmldom"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var prefixLen int

func SetPrefixLen(n int) {
	prefixLen = n
}

func Utf8ToGbk(b []byte) []byte {
	r := bytes.NewReader(b)

	decoder := transform.NewReader(r, simplifiedchinese.GBK.NewEncoder()) //GB18030

	content, _ := ioutil.ReadAll(decoder)

	return content
}

func GbkToUtf8(b []byte) []byte {
	tfr := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(tfr)
	if e != nil {
		return nil
	}
	return d
}

func GenFrame(frame []byte) []byte {
	return genFrameWithPrefix(frame, prefixLen)
}

func genFrameWithPrefix(frame []byte, prefixLen int) []byte {
	prefixLenStr := fmt.Sprintf("%d", prefixLen)
	prefix := fmt.Sprintf("%0"+prefixLenStr+"d", len(frame))
	return append([]byte(prefix), frame...)
}

func GetFrameLengthSlice(frame []byte) []byte {
	if frame == nil {
		return nil
	}

	return frame[:6]
}

func GetSeqNo(frame []byte) string {
	var document = xmldom.Must(xmldom.ParseXML(string(frame))).Root
	if node := document.QueryOne("//SEQ_NO"); node != nil {
		return node.Text
	}

	return ""
}

func UUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x%x%x%x%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

func Log_withUUID(uuid, data string) {
	Logger.Println("[" + uuid + "]" + data)
}
