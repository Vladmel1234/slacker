package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/slack-io/slacker"
)

var moodSurveyView = slack.ModalViewRequest{
	Type:       "modal",
	CallbackID: "mood-survey-callback-id",
	Title: &slack.TextBlockObject{
		Type: "plain_text",
		Text: "Which mood are you in?",
	},
	Submit: &slack.TextBlockObject{
		Type: "plain_text",
		Text: "Submit",
	},
	NotifyOnClose: true,
	Blocks: slack.Blocks{
		BlockSet: []slack.Block{
			&slack.InputBlock{
				Type:           "input",
				DispatchAction: true,
				BlockID:        "mood",
				Label: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "Input mood",
				},
				Element: &slack.PlainTextInputBlockElement{
					MaxLength: 23,
					Type:      slack.METPlainTextInput,
					ActionID:  "mood",
					Placeholder: &slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "Enter your mood",
					},
				},
			},
		},
	},
}

// Implements a basic interactive command with modal view.
func main() {
	bot := slacker.NewClient(
		os.Getenv("SLACK_BOT_TOKEN"),
		os.Getenv("SLACK_APP_TOKEN"),
		slacker.WithDebug(false),
		slacker.WithSelfAck(true),
	)

	bot.AddCommand(&slacker.CommandDefinition{
		Command: "mood",
		Handler: moodCmdHandler,
	})

	bot.AddInteraction(&slacker.InteractionDefinition{
		InteractionID: "mood-survey-callback-id",
		Handler:       moodViewHanlerWithBotClient(bot),
		Type:          slack.InteractionTypeViewSubmission,
	})

	bot.AddInteraction(&slacker.InteractionDefinition{
		InteractionID: "mood-survey-callback-id",
		Handler:       moodViewHanlerWithBotClient(bot),
		Type:          slack.InteractionTypeViewClosed,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func moodCmdHandler(ctx *slacker.CommandContext) {
	_, err := ctx.SlackClient().OpenView(
		ctx.Event().Data.(*slack.SlashCommand).TriggerID,
		moodSurveyView,
	)
	if err != nil {
		log.Printf("ERROR openEscalationModal: %v", err)
	}
}

func moodViewHanlerWithBotClient(bot *slacker.Slacker) slacker.InteractionHandler {
	return func(ctx *slacker.InteractionContext, req *socketmode.Request) {
		switch ctx.Callback().Type {
		case slack.InteractionTypeViewSubmission:
			{

				valid, errorMessage := validateInput(ctx.Callback().View.State.Values["mood"]["mood"].Value)
				if !valid {
					errorResponse := map[string]interface{}{
						"response_action": "errors",
						"errors": map[string]string{
							"mood": errorMessage, // Block ID for the input field
						},
					}

					bot.SocketModeClient().Ack(*req, errorResponse)

				} else {
					bot.SocketModeClient().Ack(*req)

				}
				viewState := ctx.Callback().View.State.Values
				fmt.Printf(
					"Mood view submitted.\nMood: %s\n",
					viewState["mood"]["mood"].SelectedOption.Value,
				)
			}
		case slack.InteractionTypeViewClosed:
			{
				fmt.Print("Mood view closed.\n")
			}
		}
	}
}

func validateInput(input string) (bool, string) {
	re := regexp.MustCompile(`^[A-Za-z0-9-]{1,23}$`)
	if !re.MatchString(input) {
		return false, "Mood can not contain special characters"
	}
	return true, ""
}
