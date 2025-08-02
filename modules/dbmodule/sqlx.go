package dbmodule

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func newSqlxDB(conf *dbConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect(conf.Driver, conf.Dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s 数据库无法访问: %s", conf.Driver, err)
	}

	db.SetMaxIdleConns(conf.MaxIdleConn)
	db.SetMaxOpenConns(conf.MaxConn)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}
