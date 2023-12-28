package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

// create an error object if a no redirect is found (404)
var ErrNotFound = errors.New(toString(http.StatusNotFound))

// create an logging for default logging -> stdout
// var stdout = slog.New(slog.NewJSONHandler(os.Stdout, nil))
var stdout = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

// create logging for errors -> stderr
var stderr = slog.New(slog.NewJSONHandler(os.Stderr, nil))

var ctx = context.Background()

// convert any value to a string
func toString(value any) string {
	return fmt.Sprintf("%v", value)
}

// convert a string to an integer
// it only supports string because Atoi only supports string
// if you want other types use the following syntax
// -> toInt(toString(errors.Error()))
func toInt(value string) int {
	val, err := strconv.Atoi(value)

	if err != nil {
		// if an error occurs: panic
		panic(err)
	}

	return val
}

// retrieve the value from an env variable
// if the key does not exist, return an fallback
func getEnv(key string, fallback string) string {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return fallback
}

// retrieve a value from redis
func getRedis(searchString string) (string, error) {
	// connect to the readis server
	rdb := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_HOST", "127.0.0.1:6379"),
		Username: getEnv("REDIS_USERNAME", "default"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       toInt(getEnv("REDIS_DB", "0")),
	})

	// retrieve a value by key
	val, err := rdb.Get(ctx, searchString).Result()

	if err == redis.Nil {
		// if redis returns nil the Key was not found
		// so we return an custom NotFound Error
		return "", ErrNotFound
	}

	if err != nil {
		// if some other error occures
		// return a normal error
		return "", err
	}

	return val, nil
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	request_uri := req.URL.RequestURI()

	if getEnv("PROXY_MODE", "false") == "true" {
		// if proxy mode is enabled, override request_uri
		request_uri = req.Header["X-Original-Uri"][0]
	}

	// get the value based on the request Host Header and the URL
	request_string := strings.TrimRight(fmt.Sprintf("%s%s", req.Host, request_uri), "/")
	val, err := getRedis(request_string)

	if err == nil {
		// Log a successfull request to stdout
		stdout.Info("redirect",
			"status", http.StatusMovedPermanently,
			"host", request_string,
			"target", val)

		if getEnv("PROXY_MODE", "false") == "true" {
			// if proxymode is enabled, return 403 forbidden
			// instead of 301 with the location header
			http.Redirect(w, req, val, http.StatusForbidden)
			return
		}

		// redirect with 301 Moved Permanently
		http.Redirect(w, req, val, http.StatusMovedPermanently)
		return
	}

	if errors.Is(err, ErrNotFound) {
		// if the custom NotFound Error is returned
		// Log the Request and return 404
		stdout.Info("not found",
			"status", http.StatusNotFound,
			"host", request_string)


		if getEnv("PROXY_MODE", "false") == "true" {
			// if proxymode is enabled, return 200 OK
			// instead of 404 so that the request is progressed normaly
			w.WriteHeader(http.StatusOK)
			return
		}

		// return 404 Not Found
		http.NotFound(w, req)
		return
	}

	// if everything else went wrong log an error to stderr
	stderr.Error(err.Error(), "status", http.StatusInternalServerError, "host", request_string)

	if getEnv("PROXY_MODE", "false") == "true" {
		// if proxymode is enabled, return 200 OK
		// even if someting went wrong, so the normal
		// application can still work
		w.WriteHeader(http.StatusOK)
		return
	}

	// and return 500 Internal Server Error
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
