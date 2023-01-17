package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"vtmmock/db"
	. "vtmmock/log"
)

func createNewRules(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	Logger.Println("[createNewRules] request :" + string(reqBody))
	var req db.Req
	if err := json.Unmarshal(reqBody, &req); err != nil {
		Logger.Println("[createNewRules]  :" + err.Error())
		w.Write([]byte(err.Error()))
		return
	}
	var c db.Cmd
	ch := make(chan db.Res)
	c.Request.Data = req.Data
	c.Request.Ty = req.Ty
	c.Request.Id = req.Id
	c.Res = ch

	db.DbCh <- c
	res := <-ch
	reply, _ := json.Marshal(&res)
	Logger.Println("[createNewRules] reply :" + string(reply))
	json.NewEncoder(w).Encode(res)
}

func createNewTemplate(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	Logger.Println("[createNewTemplate] request :" + string(reqBody))
	var req db.TemplateReq
	if err := json.Unmarshal(reqBody, &req); err != nil {
		Logger.Println("[createNewTemplate]  :" + err.Error())
		w.Write([]byte(err.Error()))
		return
	}
	ch := make(chan db.TemplateRes)
	req.Ch = ch

	db.TemplatesCh <- req
	res := <-ch

	json.NewEncoder(w).Encode(res)
}

func StartCmd(port string) {
	http.HandleFunc("/createNewRules", createNewRules)
	http.HandleFunc("/createNewTemplate", createNewTemplate)
	http.ListenAndServe(":"+port, nil)
}
