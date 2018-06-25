package repository

import (
  "encoding/json"
  "github.com/jinzhu/gorm"
  "reflect"
  "errors"
)

const (
  STATEMENT = "id = ?"
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

func (r *Repository) Delete(condition interface{}) (bool, error) {
  var (
    entity    reflect.Value
    operation *gorm.DB
  )

  if reflect.ValueOf(condition).Kind() != reflect.Ptr {
    return false, errors.New("The target struct is required to be a pointer")
  }

  entity = reflect.New(reflect.ValueOf(condition).Type().Elem()).Elem()
  operation = r.DB.Where(condition).Delete(entity.Interface())

  if operation.RecordNotFound() {
    return false, nil
  }

  if operation.Error != nil {
    return true, operation.Error
  }

  return true, nil
}

func (r *Repository) Read(ID string, entity interface{}) (bool, error) {
  if reflect.ValueOf(entity).Kind() != reflect.Ptr {
    return false, errors.New("The target struct is required to be a pointer")
  }

  var operation *gorm.DB = r.DB.First(entity, ID)

  if operation.RecordNotFound() {
    return false, nil
  }

  if operation.Error != nil {
    return true, operation.Error
  }

  return true, nil
}

func (r *Repository) ReadByConditions(entity, conditions interface{}) (bool, error) {
  if reflect.ValueOf(entity).Kind() != reflect.Ptr {
    return false, errors.New("The target struct is required to be a pointer")
  }

  var operation *gorm.DB = r.DB.First(entity, conditions)

  if operation.RecordNotFound() {
    return false, nil
  }

  if operation.Error != nil {
    return true, operation.Error
  }

  return true, nil
}

func (r *Repository) Update(condition, entity interface{}) error {
  if reflect.ValueOf(condition).Kind() != reflect.Ptr {
    return errors.New("The target struct is required to be a pointer")
  }

  if reflect.ValueOf(entity).Kind() != reflect.Ptr {
    return errors.New("The target struct is required to be a pointer")
  }

  if operation := r.DB.Model(entity).Where(condition).Updates(entity); operation.Error != nil {
    return operation.Error
  }

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
