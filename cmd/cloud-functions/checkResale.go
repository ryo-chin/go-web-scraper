package p

import (
	"context"
	"encoding/json"
	"firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
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
		return withStack(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return withStack(err)
	}
	defer client.Close()

	url := "https://grips-outdoor.jp/?pid=76851971"
	resp, err := http.Get(url)
	if err != nil {
		return withStack(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(transform.NewReader(resp.Body, japanese.EUCJP.NewDecoder()))
	if err != nil {
		return withStack(err)
	}

	var msg string
	var onSale bool
	body := string(b)
	if strings.Contains(body, "カートに入れる") {
		msg = fmt.Sprintf("販売中です!!売り切れる前にどうぞ!!\n%s", url)
		onSale = true
	} else {
		r := regexp.MustCompile(`次回の販売は\d+月末.+を予定しております。`)
		matchStrings := r.FindAllString(body, -1)
		msg = fmt.Sprintf("売り切れ中...%s", matchStrings[0])
		onSale = false
	}
	if !onSale {
		log.Println(fmt.Sprintf("exit because not on sale. msg=%s", msg))
		os.Exit(0)
	}

	docs := client.Collection("pushTokens").Limit(10).Documents(ctx)
	snapshots, err := docs.GetAll()
	if err != nil {
		return withStack(err)
	}

	// Slack WebHook通知
	err = pushToSlack(msg)
	if err != nil {
		log.Println(fmt.Sprintf("slack push is failed. err=%+v", err))
	}

	// FCMプッシュ通知
	var msgs []*messaging.Message
	for _, docsnap := range snapshots {
		type PushToken struct {
			Token string `firestore:"token"`
		}
		var pushToken PushToken
		if err := docsnap.DataTo(&pushToken); err != nil {
			return withStack(err)
		}
		webpush := new(messaging.WebpushConfig)
		webpush.Notification = &messaging.WebpushNotification{
			Title: "再販売通知",
			Body:  msg,
		}
		webpush.FcmOptions = &messaging.WebpushFcmOptions{
			Link: url,
		}
		msgs = append(msgs, &messaging.Message{
			Token:   pushToken.Token,
			Webpush: webpush,
		})
	}

	fcmService, err := app.Messaging(ctx)
	if err != nil {
		return withStack(err)
	}
	res, err := fcmService.SendAll(ctx, msgs)
	if err != nil {
		return withStack(err)
	}
	handlePushResponse(res)
	return nil
}

func handlePushResponse(r *messaging.BatchResponse) {
	log.Println(fmt.Sprintf("send push result {success=%d, failure=%d}", r.SuccessCount, r.FailureCount))
}

func withStack(err error) error {
	log.Println(fmt.Sprintf("%+v", errors.WithStack(err)))
	return err
}

func InitFirebase(pID string, ctx context.Context) (*firebase.App, error) {
	conf := &firebase.Config{ProjectID: pID}
	app, err := firebase.NewApp(ctx, conf)
	return app, err
}

func pushToSlack(msg string) (err error) {
	webhookURL := os.Getenv("RESALE_SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return errors.New("WebHookURL is not exists")
	}
	p, err := json.Marshal(SlackMessage{Text: msg})
	if err != nil {
		return err
	}
	resp, err := http.PostForm(webhookURL, url.Values{"payload": {string(p)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type SlackMessage struct {
	Text string `json:"text"`
}
