// Filename: cmd/api/main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"sync"
	"time"

	"coltech.osborncollins.net/internal/data"
	"coltech.osborncollins.net/internal/jsonlog"
	"coltech.osborncollins.net/internal/mailer"
	_ "github.com/lib/pq"
)

// The Application Version Number
const version = "2.0.0"

// The Configuration Settings
type config struct {
	port int
	env  string // Development, Staging, Production, ETC.
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// Dependency Injection
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	// Read in flags that are needed to populate our config
	flag.IntVar(&cfg.port, "port", 4000, "API Server Port") // When using a struct we must use IntVar
	flag.StringVar(&cfg.env, "env", "development", "Environment( Development | Staging | Production )")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("COLTECH_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	// These are flags for rate limiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum request per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst per second")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enabled the rate limiter")
	// These are the flags for the mailer
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "30ebafc237533c", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "ce73e749caad00", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Coltech <no-reply@coltech.osborncollins.net>", "SMTP sender")

	flag.Parse()

	//Create a logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	// Create the connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// If anything happens we would like to close connection
	defer db.Close()
	//Log the sucessful connection pool
	logger.PrintInfo("Database connection pool established", nil)

	//Create an instance of our application struct
	// We are using the application struct for dependecy injection
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	// Call app.serve() to start the server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}

//The openDB() function returns a pointer to an sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	// Test the connection pool
	// Create a conteext with a 5 second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
