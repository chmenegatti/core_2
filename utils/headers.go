package utils

const (
  TRANSACTION_ID  = "x-transaction-id"
  API_VERSION	  = "x-api-version"
  API_ACTION	  = "x-api-action"
)

type Headers struct {
  TransactionID	string
  Version	string
  Action	string
}

func GetHeader(infos map[string]interface{}) Headers {
  var h Headers

  h.TransactionID = getHeader(TRANSACTION_ID, infos)
  h.Version = getHeader(API_VERSION, infos)
  h.Action = getHeader(API_ACTION, infos)

  return h
}

func getHeader(key string, infos map[string]interface{}) string {
  var s string

  if infos != nil {
    if v, ok := infos[key]; ok {
      s = v.(string)
    }
  }

  return s
}
