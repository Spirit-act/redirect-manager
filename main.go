package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New(toString(http.StatusNotFound))

var stdout = slog.New(slog.NewJSONHandler(os.Stdout, nil))
var stderr = slog.New(slog.NewJSONHandler(os.Stderr, nil))

var ctx = context.Background()

func toString(value any) string {
	return fmt.Sprintf("%v", value)
}

func getRedis(host_header string) (string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
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
		stdout.Info("redirect", "status", "301", "host", req.Host, "target", val)
		http.Redirect(w, req, val, http.StatusMovedPermanently)
		return
	}

	if errors.Is(err, ErrNotFound) {
		stdout.Info("not found", "status", "404", "host", req.Host)
		http.NotFound(w, req)
		return
	}

	stderr.Error(err.Error(), "status", "500", "host", req.Host)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":8090", nil)
}
