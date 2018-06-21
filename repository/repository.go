package repository

import (
  "encoding/json"
  "github.com/jinzhu/gorm"
  //"github.com/satori/go.uuid"
  "reflect"
  "errors"
)

type Repository struct {
  DB  *gorm.DB
}

func (r *Repository) Create(src, dst interface{}) error {
  var err error

  if reflect.ValueOf(dst).Kind() != reflect.Ptr {
    return errors.New("The target struct is required to be a pointer")
  }

  if err = r.marshalAndUnmarshal(src, dst); err != nil {
    return err
  }

  if err = r.DB.Create(dst).Error; err != nil {
    return err
  }

  return nil
}

func (r *Repository) Delete(id string) error {
  return nil
}

func (r *Repository) marshalAndUnmarshal(src, dst interface{}) error {
  var (
    body  []byte
    err	  error
  )

  if body, err = json.Marshal(src); err != nil {
    return err
  }

  if err = json.Unmarshal(body, dst); err != nil {
    return err
  }

  return nil
}
