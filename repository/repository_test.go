package repository

import (
	"fmt"
	"testing"
	"database/sql/driver"
	"github.com/jinzhu/gorm"
	"github.com/erikstmartin/go-testdb"
)

type Src struct {
	ID	string  `json:"ID,omitempty"`
	Name	string  `json:"Name,omitempty"`
	NameID	string  `json:"NameID,omitempty"`
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
		s   = Src{}
		r   Repository
		err error
	)

	testdb.SetQueryWithArgsFunc(func(query string, args []driver.Value) (driver.Rows, error) {
		columns := []string{"name", "name_id"}
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
