package config

import (
  "fmt"
  "time"

  "git-devops.totvs.com.br/intera/core/log"
  _ "github.com/go-sql-driver/mysql"
  "github.com/jinzhu/gorm"
)

const (
  DBTemplate = "%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=%s"
)

func LoadDB() (*gorm.DB, error) {
  var (
    db	*gorm.DB
    err	error
  )

  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", INIT)

  if db, err = gorm.Open("mysql", fmt.Sprintf(DBTemplate, EnvDB.Username, EnvDB.Password, EnvDB.Host, EnvDB.Port, EnvDB.DBName, EnvDB.Timeout)); err != nil {
    return db, err
  }

  db.LogMode(EnvDB.Debug)
  db.DB().SetMaxIdleConns(EnvDB.ConnsMaxIdle)
  db.DB().SetMaxOpenConns(EnvDB.ConnsMaxOpen)
  db.DB().SetConnMaxLifetime(time.Duration(EnvDB.ConnsMaxLifetime))

  return db, nil
}
