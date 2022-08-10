package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ysmood/got/lib/gop"
)

type Rsp struct {
	Rows  []TotalItem `json:"Rows"`
	Total int         `json:"Total"`
}

type TotalItem struct {
	Faid        string `json:"faid"`
	Fabh        string `json:"fabh"`
	Famc        string `json:"famc"`
	Postinfoid  string `json:"postinfoid"`
	Postinfono  string `json:"postinfono"`
	Takejob     string `json:"takejob"`
	Takenum     string `json:"takenum"`
	Companyid   string `json:"companyid"`
	Companyname string `json:"companyname"`
	Dshrs       string `json:"dshrs"`
	Shtgyjf     string `json:"shtgyjf"`
	Shtgwjf     string `json:"shtgwjf"`
	Shbtg       string `json:"shbtg"`
	Zbmrs       string `json:"zbmrs"`
	Tjsj        string `json:"tjsj"`
}

var codes = map[string][]string{}

func index(c *gin.Context) {
	computer := c.Query("computer")
	architecture := c.Query("architecture")
	media := c.Query("media")
	if computer == "" && architecture == "" && media == "" { //全空默认为全选
		computer = "on"
		architecture = "on"
		media = "on"
	}
	query := map[string]string{
		"computer":     computer,
		"architecture": architecture,
		"media":        media,
	}
	var ids []string
	for k, v := range query {
		if v != "on" {
			continue
		}
		ids = append(ids, codes[k]...)
	}
	gop.P(query)
	totalURL := `https://app.hrss.xm.gov.cn/syzp/index/index/jobApply!findJobApplyTotal?&faid=10002789`
	detailURL := `https://app.hrss.xm.gov.cn/syzp/index/postinfo!viewPostinfo4index?postinfoid=`
	data, err := getList(totalURL)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	if data.Total == 0 {
		c.String(http.StatusOK, "list empty")
		return
	}
	data, err = getList(totalURL + "&pagesize=" + fmt.Sprint(data.Total))
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	if data.Total == 0 {
		c.String(http.StatusOK, "list empty")
		return
	}
	ruleList := map[string]TotalItem{}
	for _, row := range data.Rows {
		ruleList[row.Postinfono] = row
	}
	data.Rows = filterList(ids, ruleList)
	data.Total = len(data.Rows)
	gop.P(data.Total)
	c.HTML(http.StatusOK, "index.html", map[string]interface{}{
		"computer":     computer,
		"architecture": architecture,
		"media":        media,
		"data":         data,
		"detailURL":    detailURL,
	})
}

func filterList(ids []string, list map[string]TotalItem) []TotalItem {
	res := make([]TotalItem, 0, len(ids))
	for _, id := range ids {
		if d, ok := list[id]; ok {
			res = append(res, d)
		}
	}
	return res
}

func getList(totalURL string) (Rsp, error) {
	rsp, err := http.Get(totalURL)
	if err != nil {
		return Rsp{}, errors.New("get detail" + err.Error())
	}
	defer rsp.Body.Close()
	b, _ := ioutil.ReadAll(rsp.Body)
	//b, _ := ioutil.ReadFile("./data.json")
	b = bytes.ReplaceAll(b, []byte("Rows"), []byte(`"Rows"`))
	b = bytes.ReplaceAll(b, []byte("Total"), []byte(`"Total"`))
	data := Rsp{}
	_ = json.Unmarshal(b, &data)
	return data, nil
}

func main() {
	b, _ := ioutil.ReadFile("config/data.json")
	_ = json.Unmarshal(b, &codes) //init config
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"add": func(v int) int {
			return v + 1
		},
	})
	r.LoadHTMLFiles("./index.html")
	r.StaticFS("/layui", http.Dir("./layui"))
	r.GET("/", index)
	r.Run(":80") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
