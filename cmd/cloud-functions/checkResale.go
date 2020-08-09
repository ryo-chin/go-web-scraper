package p

import (
	"context"
	"encoding/json"
	"firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func CheckResale(w http.ResponseWriter, r *http.Request) {
	// Use the application default credentials
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "github-api-app-2acb5"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

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
	var msg string
	body := string(b)
	if strings.Contains(body, "カートに入れる") {
		msg = fmt.Sprintf("url= %s は販売中だよ", url)
		response = d{true, msg}
	} else {
		r := regexp.MustCompile(`次回の販売は\d+月末.+を予定しております。`)
		matchStrings := r.FindAllString(body, -1)
		msg = fmt.Sprintf("url= %s は売り切れ中...%s", url, matchStrings[0])
		response = d{false, msg}
	}
	e, err := json.Marshal(response)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	doc := client.Collection("pushTokens").Doc("1")
	log.Println(*doc)
	docsnap, err := doc.Get(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	type PushToken struct {
		Token string `firestore:"token"`
	}
	var pushToken PushToken
	if err := docsnap.DataTo(&pushToken); err != nil {
		log.Fatalln(err)
	}

	fcmService, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	webpush := new(messaging.WebpushConfig)
	webpush.Notification = &messaging.WebpushNotification{
		Title: "再販売通知",
		Body:  msg,
	}
	_, err = fcmService.Send(ctx, &messaging.Message{
		Token:   pushToken.Token,
		Webpush: webpush,
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprint(w, string(e))
}
