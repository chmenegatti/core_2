package config

import (
	"fmt"
	"time"

	"gitlab.com/ascenty/core/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const (
	DBTemplate = "%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=%s"
)

func LoadDB() (map[string]*gorm.DB, error) {
	var (
		mdb	= make(map[string]*gorm.DB)
		err	error
	)

	EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadDB", INIT)

	for key, values := range EnvDB {
		var db *gorm.DB

		if db, err = gorm.Open("mysql", fmt.Sprintf(DBTemplate, values.Username, values.Password, values.Host, values.Port, values.DBName, values.Timeout)); err != nil {
			return mdb, err
		}

		db.LogMode(values.Debug)
		db.DB().SetMaxIdleConns(values.ConnsMaxIdle)
		db.DB().SetMaxOpenConns(values.ConnsMaxOpen)
		db.DB().SetConnMaxLifetime(time.Duration(values.ConnsMaxLifetime))

		mdb[key] = db

	}

	return mdb, nil
}
