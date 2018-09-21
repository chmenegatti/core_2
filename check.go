package core

import (
  "gitlab-devops.totvs.com.br/microservices/core/config"

  "net/http"
  "strings"
  "fmt"
)

const (
  PORT_DEFAULT = "80"
)

func healthCheck() error {
  var (
    mux	  *http.ServeMux
    url	  = config.EnvConfig.CheckURL
    port  string
  )

  if strings.Split(url, "/")[0] != "" {
    url = "/" + url
  }

  if len(strings.Split(url, ":")) > 1 {
    port = strings.Split(url, ":")[1]
    url = strings.Split(url, ":")[0]
  } else {
    port = PORT_DEFAULT
  }

  mux = http.NewServeMux()
  mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "Hello!!!\n")
  })

  return http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
}
