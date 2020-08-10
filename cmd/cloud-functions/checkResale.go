package p

import (
	"context"
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

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

func CheckResale(ctx context.Context, m PubSubMessage) error {
	// Use the application default credentials
	app, err := InitFirebase("github-api-app-2acb5", ctx)
	if err != nil {
		return err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	url := "https://grips-outdoor.jp/?pid=76851971"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(transform.NewReader(resp.Body, japanese.EUCJP.NewDecoder()))
	if err != nil {
		return err
	}

	var msg string
	body := string(b)
	if strings.Contains(body, "カートに入れる") {
		msg = fmt.Sprintf("url= %s は販売中だよ", url)
	} else {
		r := regexp.MustCompile(`次回の販売は\d+月末.+を予定しております。`)
		matchStrings := r.FindAllString(body, -1)
		msg = fmt.Sprintf("url= %s は売り切れ中...%s", url, matchStrings[0])
	}

	doc := client.Collection("pushTokens").Doc("1")
	log.Println(*doc)
	docsnap, err := doc.Get(ctx)
	if err != nil {
		return err
	}
	type PushToken struct {
		Token string `firestore:"token"`
	}
	var pushToken PushToken
	if err := docsnap.DataTo(&pushToken); err != nil {
		return err
	}

	fcmService, err := app.Messaging(ctx)
	if err != nil {
		return err
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
		return err
	}

	log.Println(string(m.Data))
	return nil
}

func InitFirebase(pID string, ctx context.Context) (*firebase.App, error) {
	conf := &firebase.Config{ProjectID: pID}
	app, err := firebase.NewApp(ctx, conf)
	return app, err
}
