package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const base_format = "2006-01-02 15:04:05"

const (
	dbDriverName = "sqlite3"
	dbName       = "./data.db3"
)

type wxfile struct {
	filename  string
	checktime time.Time
}

func initDb() (*sql.DB, error) {
	db, err := sql.Open(dbDriverName, dbName)
	if err != nil {
		return nil, err
	}

	err = createTable(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// 创建空的数据表
func createTable(db *sql.DB) error {
	sql := `create table if not exists "wxfile" (
		"id" integer primary key autoincrement,
		"filename" text not null,
		"checkintime" text not null
	)`
	_, err := db.Exec(sql)
	return err
}

// 返回的行数
func checkCount(rows *sql.Rows) (count int) {
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			fmt.Println(err)
			return 0
		}
	}
	return count
}

// 记录处理过的文件
func insertData(db *sql.DB, f string) error {
	sql := `select count(id) from wxfile where filename = ?`

	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}

	rows, err := stmt.Query(f)
	if err != nil {
		return err
	}

	count := checkCount(rows)
	if count > 0 {
		fmt.Println("rows:", count)
		rows.Close()
		return nil
	}

	fmt.Println("insert data:", f)
	sql = `insert into wxfile(filename, checkintime) values(?,?)`
	stmt, err = db.Prepare(sql)
	if err != nil {
		return err
	}

	nt := time.Now()
	s_time := nt.Format(base_format)
	_, err = stmt.Exec(f, s_time)
	return err

}
