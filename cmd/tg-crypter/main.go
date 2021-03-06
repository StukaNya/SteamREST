package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/StukaNya/TgCrypter/api"
	store "github.com/StukaNya/TgCrypter/storage/user"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "./config.yml", "path to config file")
}

func main() {
	// Init context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init OS sigquit channel
	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, os.Interrupt, syscall.SIGINT)

	go func() {
		oscall := <-sigquit
		log.Printf("syscall:%+v", oscall)
		cancel()
	}()

	// Load config from YAML file
	config := NewConfig()
	err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}

	// Configure logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.Logger.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.SetLevel(level)

	// Configure DB and store layer
	dbURL := config.DbConfig.DatabaseURL
	db, err := newDB(dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// User storage
	userSt := store.NewUserStorage(db)
	err = userSt.InitTable(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Server startup
	server := api.NewAPIServer(logger, &config.ServerConfig, userSt)
	err = server.Serve(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// Open database
func newDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return db, err
	}

	return db, nil
}
