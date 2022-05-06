package repository

import (
	"fmt"
	"bytes"
	"errors"
	"testing"
	"database/sql/driver"
	"github.com/jinzhu/gorm"
	"github.com/erikstmartin/go-testdb"
)

type JSON []byte
func (j JSON) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		errors.New("Invalid Scan Source")
	}
	*j = append((*j)[0:0], s...)
	return nil
}
func (m JSON) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}
func (m *JSON) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("null point exception")
	}
	*m = append((*m)[0:0], data...)
	return nil
}
func (j JSON) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}
func (j JSON) Equals(j1 JSON) bool {
	return bytes.Equal([]byte(j), []byte(j1))
}

type Src struct {
	ID	    string	    `json:"ID,omitempty"`
	Name	    string	    `json:"Name,omitempty"`
	NameID	    string	    `json:"NameID,omitempty"`
	Attributes  string	    `json:"Attributes,omitempty" gorm:"type:json"`
}

type Dst struct {
	Name  string  `json:"Name,omitempty"`
	ID    string  `json:"NameID,omitempty"`
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
		d   = Dst{}
		r   Repository
		err error
	)

	testdb.SetExecWithArgsFunc(func(query string, args []driver.Value) (driver.Result, error) {
		fmt.Println("query: ", query)
		fmt.Println("args: ", args)

		return testdb.NewResult(1, nil, 1, nil), nil
	})

	r.DB = loadDB()
	d = Dst{Name: "test", ID: "1"}

	if err = r.Create(&d); err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_Delete(t *testing.T) {
	var (
		s   = Src{}
		r   Repository
		err error
	)

	testdb.SetExecWithArgsFunc(func(query string, args []driver.Value) (driver.Result, error) {
		fmt.Println("query: ", query)

		return testdb.NewResult(1, nil, 1, nil), nil
	})

	r.DB = loadDB()
	s = Src{Name: "test"}

	if _, err = r.Delete(&s); err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_Read(t *testing.T) {
	var (
		s	= Src{ID: "1"}
		r	Repository
		err	error
	)

	testdb.SetQueryWithArgsFunc(func(query string, args []driver.Value) (driver.Rows, error) {
		columns := []string{"name", "name_id"}
		rows := "Name,NameID"

		fmt.Println("query: ", query)
		fmt.Println("args: ", args)

		return testdb.RowsFromCSVString(columns, rows), nil
	})

	r.DB = loadDB()

	if _, err = r.Read(s.ID, &s); err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Println(s)
}

func Test_ReadByConditions(t *testing.T) {
	var (
		s   = []Src{}
		r   Repository
		err error
	)

	testdb.SetQueryWithArgsFunc(func(query string, args []driver.Value) (driver.Rows, error) {
		columns := []string{"name", "name_id", "attributes"}
		rows := "Name,NameID"

		fmt.Println("query: ", query)
		fmt.Println("args: ", args)

		return testdb.RowsFromCSVString(columns, rows), nil
	})

	r.DB = loadDB()

	if _, err = r.ReadByConditions(&s, Src{Name: "Name"}); err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Println(s)
}

func Test_Update(t *testing.T) {
	var (
		s   = Src{}
		r   Repository
		err error
	)

	testdb.SetExecWithArgsFunc(func(query string, args []driver.Value) (driver.Result, error) {
		fmt.Println("query: ", query)
		fmt.Println("args: ", args)

		return testdb.NewResult(1, nil, 1, nil), nil
	})

	r.DB = loadDB()
	s = Src{Name: "Name"}

	if err = r.Update(&Src{ID: "1", Name: "Name2"}, &s); err != nil {
		t.Fatalf(err.Error())
	}
}
