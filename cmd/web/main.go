package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

func main() {
	// connect to DB
	db, err := initDB()
	if err != nil {
		log.Println(err)
	}
	db.Ping()

	// create seasions
	session := initSession()

	// create loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create chanels

	// create waitGroups
	wg := sync.WaitGroup{}

	// set Up app config

	app := Config{
		Session:  session,
		DB:       db,
		Wait:     &wg,
		InfoLog:  infoLog,
		ErrorLog: errLog,
	}

	// set up mail

	// listen for web connections
	err = app.serve()
	if err != nil {
		log.Println(err)
	}
}

func (app *Config) serve() error {
	// start http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	app.InfoLog.Println("Starting web server...")
	err := srv.ListenAndServe()
	if err != nil {
		app.ErrorLog.Println("Error starting web server!")
		return err
	}
	return nil
}

func initDB() (*sql.DB, error) {
	conn := connectToDB()
	if conn == nil {
		return nil, errors.New("can't connect to database")
	}
	return conn, nil
}

func connectToDB() *sql.DB {
	counts := 0

	// dsn - DataSourceName
	dsn := "host=localhost port=5432 user=postgres password=password dbname=concurrency sslmode=disable timezone=UTC connect_timeout=5" // TODO Change

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not yet ready...")
		} else {
			log.Println("connected to database!")
			return connection
		}

		if counts > 10 {
			return nil
		}
		log.Println("Backing off for 1 second")
		time.Sleep(1 * time.Second)
		counts++
		continue
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func initSession() *scs.SessionManager {
	// set up session
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}

func initRedis() *redis.Pool {
	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}
	return redisPool
}
