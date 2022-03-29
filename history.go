package core

import (
	"encoding/json"
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
				"Content-Type":	"application/json"
		        },
		},
	)

	path = fmt.Sprintf("%s/v1/elastic/documents", config.EnvConfig.TotvsGatewayHistoryUrl)

	if output, err = httpRequest.Request(httpRequest.POST, path, options); err != nil {
		return
	}

	if output.Code != 201 {
		var msg Message

		if err = json.Unmarshal(output.Body, &msg); err != nil {
			msg = string(output.Body)
		}

		err = errors.New(msg)
	}

	return
}
