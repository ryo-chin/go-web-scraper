package main

import (
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func main() {
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
	body := string(b)
	if strings.Contains(body, "カートに入れる") {
		fmt.Println(fmt.Sprintf("url= %s は販売中だよ", url))
	} else {
		r := regexp.MustCompile(`次回の販売は\d+月末頃を予定しております。`)
		matchStrings := r.FindAllString(body, -1)
		fmt.Println(fmt.Sprintf("url= %s は売り切れ中...%s", url, matchStrings[0]))
	}
}
