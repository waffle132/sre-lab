package main

import (
    "database/sql"
    "log"

    _ "modernc.org/sqlite"
)

func main() {
    db, err := sql.Open("sqlite", "runtime/go-practice.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    log.Println("数据库连接成功")
}