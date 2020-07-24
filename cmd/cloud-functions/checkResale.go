package p

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func CheckResale(w http.ResponseWriter, r *http.Request) {
	url := "https://grips-outdoor.jp/?pid=76851971"
	resp, err := http.Get(url)
	if err != nil {
		println(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(transform.NewReader(resp.Body, japanese.EUCJP.NewDecoder()))
	if err != nil {
		panic(err)
	}

	type d struct {
		OnSale  bool   `json:"onSale"`
		Message string `json:"message"`
	}
	var response d
	body := string(b)
	if strings.Contains(body, "カートに入れる") {
		response = d{true, fmt.Sprintf("url= %s は販売中だよ", url)}
	} else {
		r := regexp.MustCompile(`次回の販売は\d+月末頃を予定しております。`)
		matchStrings := r.FindAllString(body, -1)
		response = d{false, fmt.Sprintf("url= %s は売り切れ中...%s", url, matchStrings[0])}
	}
	e, err := json.Marshal(response)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, string(e))
}
