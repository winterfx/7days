package test

import (
	"database/sql"
	"fmt"
	"geeorm"

	//导入github.com/mattn/go-sqlite3包。
	//不直接在代码中引用该包的任何公共成员。
	//允许包的初始化代码被执行，这里主要是注册SQLite3的数据库驱动。
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func TestConnect(t *testing.T) {
	db, err := sql.Open("sqlite3", "gee_test.db")
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		db.Close()
	}()
	_, _ = db.Exec("DROP TABLE IF EXISTS User;")
	_, _ = db.Exec("CREATE TABLE User(Name text);")
	result, err := db.Exec("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam")
	if err == nil {
		affected, _ := result.RowsAffected()
		fmt.Println(affected)
	}
	row := db.QueryRow("SELECT Name FROM User LIMIT 1")
	var name string
	if err := row.Scan(&name); err == nil {
		fmt.Println(name)
	}
}
func TestGeeOrm(t *testing.T) {
	e := geeORM.NewEngine("sqlite3", "gee.db")
	s := e.NewSession()
	r, err := s.Raw("INSERT INTO User(`Name`) values (?)", "Winter").Exec()
	if err != nil {
		t.Logf(err.Error())
		t.FailNow()
	}
	count, _ := r.RowsAffected()
	fmt.Printf("Exec success, %d affected\n", count)
}
