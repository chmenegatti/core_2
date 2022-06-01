package core

import (
	"encoding/json"
	"errors"
	"bytes"
	"fmt"

	"gitlab.com/ascenty/core/log"
	"gitlab.com/ascenty/core/config"
	"github.com/slack-go/slack"
	"gitlab.com/ascenty/httpRequestClient"
)

type TGHMessage struct {
	Message	string	`json:"message,omitempty"`
}

func publishSlack(payload, message string) {
	if config.EnvSingletons.Slack != nil {
		var (
		        err     error
		)

		var attachment = slack.Attachment{
			Text:	fmt.Sprintf("Erro ao inserir arquivo no Data-Collector - %s", config.EnvConfig.Edge),
			Color:	"#cc0000",
		        Fields: []slack.AttachmentField{{
		                Title:  payload,
		                Value:  message,
		        }},
		}

		if _, _, err = config.EnvSingletons.Slack.PostMessage(config.EnvConfig.SlackChannel, slack.MsgOptionAttachments(attachment)); err != nil {
		        config.EnvSingletons.Logger.Errorf(log.TEMPLATE_CORE, "", PACKAGE, "TGHInsertDocument", "Slack", err.Error())
		}
	}
}

func TGHInsertDocument(payload interface{}) (err error) {
	if !config.EnvConfig.EnableDataCollector {
		return
	}

	var (
		body	[]byte
		output	httpRequest.Response
		options	httpRequest.ReqOptions
		path	string
	)

	defer func() {
		if err != nil {
			publishSlack(string(body), err.Error())
		}
	}()

	if body, err = json.Marshal(payload); err != nil {
		return
	}

	options = httpRequest.SetOptions(
		httpRequest.ReqOptions{
		        PostBody: bytes.NewReader(body),
		        Headers:  map[string]string{
				"Content-Type":	"application/json",
		        },
		},
	)

	path = fmt.Sprintf("%s/v1/tgh", config.EnvConfig.TotvsGatewayHistoryUrl)

	if output, err = httpRequest.Request(httpRequest.POST, path, options); err != nil {
		return
	}

	if output.Code != 201 {
		var msg TGHMessage

		if err = json.Unmarshal(output.Body, &msg); err != nil {
			msg.Message = string(output.Body)
		}

		err = errors.New(msg.Message)
	}

	return
}
