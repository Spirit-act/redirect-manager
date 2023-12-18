# Redirect Manager

## Requirements

- go 1.21.4 or greater
- redis

## Redis

|Key|Value|
| --- | --- |
|sub.domain.tld|https://domain.tld/xyz|

## Configuration

|Key|default|info|
| --- | --- | --- |
| LISTEN_ADDR | | leave empty for IPv4 and IPv6 |
| LISTEN_PORT | 8090 | |
| REDIS_HOST | 127.0.0.1:6379 | required |
| REDIS_DB | 0 | |
| REDIS_USERNAME | default | only required with ACL |
| REDIS_PASSWORD | | optional |