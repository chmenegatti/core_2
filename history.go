package core

import (
	"encoding/json"
	"errors"
	"bytes"
	"fmt"

	"gitlab.com/ascenty/core/config"
	"git-devops.totvs.com.br/intera/httpRequestClient"
)

type TGHMessage struct {
	Message	string	`json:"message,omitempty"`
}

func TGHInsertDocument(payload interface{}) (err error) {
	var (
		body	[]byte
		output	httpRequest.Response
		options	httpRequest.ReqOptions
		path	string
	)

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
