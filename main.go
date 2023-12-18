package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New(toString(http.StatusNotFound))

var stdout = slog.New(slog.NewJSONHandler(os.Stdout, nil))
var stderr = slog.New(slog.NewJSONHandler(os.Stderr, nil))

var ctx = context.Background()

func toString(value any) string {
	return fmt.Sprintf("%v", value)
}

func getEnv(key string, fallback string) string {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return fallback
}

func getRedis(host_header string) (string, error) {
	redis_db, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_HOST", "127.0.0.1:6379"),
		Username: getEnv("REDIS_USERNAME", "default"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       redis_db,
	})

	val, err := rdb.Get(ctx, host_header).Result()

	if err == redis.Nil {
		return "", ErrNotFound
	}

	if err != nil {
		return "", err
	}

	return val, nil
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	val, err := getRedis(req.Host)

	if err == nil {
		stdout.Info("redirect",
			"status", http.StatusMovedPermanently,
			"host", req.Host,
			"target", val)
		http.Redirect(w, req, val, http.StatusMovedPermanently)
		return
	}

	if errors.Is(err, ErrNotFound) {
		stdout.Info("not found",
			"status", http.StatusNotFound,
			"host", req.Host)
		http.NotFound(w, req)
		return
	}

	stderr.Error(err.Error(), "status", http.StatusInternalServerError, "host", req.Host)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(
		fmt.Sprintf(
			"%v:%v",
			os.Getenv("LISTEN_ADDR"),
			getEnv("LISTEN_PORT", "8090")),
		nil)
}
