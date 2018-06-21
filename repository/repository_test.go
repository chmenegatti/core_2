package repository

import (
  "fmt"
  "testing"
  "database/sql/driver"
  "github.com/jinzhu/gorm"
  "github.com/erikstmartin/go-testdb"
)

type Src struct {
  Name	  string  `json:"Name,omitempty"`
  NameID  string  `json:"NameID,omitempty"`
}

type Dst struct {
  Name	string  `json:"Name,omitempty"`
  ID	string  `json:"NameID,omitempty"`
}

func loadDB() *gorm.DB {
  db, err := gorm.Open("testdb", "")
  if err != nil {
    panic(err.Error())
  }

  return db
}

func Test_Create(t *testing.T) {
  var (
    s		= Src{}
    d		= Dst{}
    r		Repository
    err		error
  )

  testdb.SetExecWithArgsFunc(func(query string, args []driver.Value) (driver.Result, error) {
    fmt.Println("query: ", query)
    fmt.Println("args: ", args)

    return testdb.NewResult(1, nil, 1, nil), nil
  })

  r.DB = loadDB()
  s = Src{Name: "test", NameID: "1"}

  if err = r.Create(s, &d); err != nil {
    t.Fatalf(err.Error())
  }
}
