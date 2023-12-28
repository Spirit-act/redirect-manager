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
| PROXY_MODE | false | if app is behind nginx auth_request |

## Proxy Mode

### Nginx Config

```
server {
    listen 8080;
    server_name _;

    location / {
        auth_request /<some_long_request_path>;
        auth_request_set $location $upstream_http_location;

        # if the auth_request returns 403 trigger the redirect
        error_page 403 = @redirection;
    }

    location @redirection {
        # handle redirection with the 
        # location header from the upstream
        return 301 $location;
    }

    location = /<some_long_request_path> {
        internal; # only internal request are allowed, external requests get 404
        proxy_pass http://app:8090; # adjust backend upstream to the redirect manager app
        proxy_pass_request_body off; # we do not need the body
        proxy_set_header Content-Length ""; # we do not need a content-header
        proxy_set_header X-Original-URI $request_uri; # repack Request URI as a Header
        proxy_set_header Host $host; # set the host header for the upstream
    }
}
```

### with caching

```
proxy_cache_path /tmp/auth_proxy_cache levels=1:2 keys_zone=auth_cache:100m max_size=1G inactive=3d;

server {
    listen 8080;
    server_name _;

    location / {
        auth_request /<some_long_request_path>;
        auth_request_set $location $upstream_http_location;

        # if the auth_request returns 403 trigger the redirect
        error_page 403 = @redirection;
    }

    location @redirection {
        # handle redirection with the 
        # location header from the upstream
        return 301 $location;
    }

    location = /<some_long_request_path> {
        internal; # only internal request are allowed, external requests get 404

        proxy_cache_valid 200 403 2d;
        proxy_cache auth_cache;
        proxy_cache_methods GET HEAD;
        proxy_cache_key $host$request_uri;

        proxy_pass http://app:8090; # adjust backend upstream to the redirect manager app
        proxy_pass_request_body off; # we do not need the body
        proxy_set_header Content-Length ""; # we do not need a content-header
        proxy_set_header X-Original-URI $request_uri; # repack Request URI as a Header
        proxy_set_header Host $host; # set the host header for the upstream
    }
}
```
