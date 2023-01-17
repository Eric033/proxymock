package helper

import "testing"

func TestGetSeqNoFound(t *testing.T) {
	d := []byte(`
	<testsuites>
		<aaa>
			<testcase classname="go-xmldom" id="ExampleParseXML" time="0.004"></testcase>
			<testcase classname="go-xmldom" id="ExampleParse" time="0.005"></testcase>
		</aaa>
		<SEQ_NO>1234567890</SEQ_NO>
	</testsuites>`)

	if GetSeqNo(d) != "1234567890" {
		t.Error(d)
	}
}

func TestGetSeqNoNotFound(t *testing.T) {
	d := []byte(`
	<testsuites>
		<aaa>
			<testcase classname="go-xmldom" id="ExampleParseXML" time="0.004"></testcase>
			<testcase classname="go-xmldom" id="ExampleParse" time="0.005"></testcase>
		</aaa>
		<SEQ_NO1>1234567890</SEQ_NO1>
	</testsuites>`)

	if GetSeqNo(d) != "" {
		t.Error()
	}
}

func TestUU(t *testing.T) {
	t.Error(UUID())
}
