package main

import (
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
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
	if strings.Contains(string(b), "カートに入れる") {
		fmt.Println(fmt.Sprintf("url= %s は販売中だよ", url))
	} else {
		fmt.Println(fmt.Sprintf("url= %s は売り切れ中...", url))
	}
}
