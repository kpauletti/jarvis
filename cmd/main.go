package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {

	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))

	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:

				switch event.Type {

				case socketmode.EventTypeEventsAPI:

					eventsAPI, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPI: %v\n", event)
						continue
					}

					socketClient.Ack(*event.Request)

					err := HandleAppMentionEventToBot(eventsAPI, client)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}(ctx, client, socketClient)

	socketClient.Run()
}

func HandleAppMentionEventToBot(event *slackevents.AppMentionEvent, client *slack.Client) error {

	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}

	text := strings.ToLower(event.Text)

	attachment := slack.Attachment{}

	if strings.Contains(text, "hello") || strings.Contains(text, "hi") {
		attachment.Text = fmt.Sprintf("Hello %s", user.Name)
		attachment.Color = "#4af030"
	} else if strings.Contains(text, "weather") {
		attachment.Text = fmt.Sprintf("Weather is sunny today. %s", user.Name)
		attachment.Color = "#4af030"
	} else {
		attachment.Text = fmt.Sprintf("I am good. How are you %s?", user.Name)
		attachment.Color = "#4af030"
	}
	_, _, err = client.PostMessage(event.Channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}
