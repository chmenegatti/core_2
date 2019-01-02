package core

import (
  "git-devops.totvs.com.br/intera/core/config"

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

  if config.EnvConfig.CheckPort != "" {
    port = config.EnvConfig.CheckPort
  } else {
    port = PORT_DEFAULT
  }

  mux = http.NewServeMux()
  mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "Hello!!!\n")
  })

  return http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
}
