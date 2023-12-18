FROM golang:1.21 AS build

WORKDIR /app

# copy mod and sum file
COPY go.mod go.sum ./
# download dependencies
RUN go mod download

# copy all go files
COPY *.go ./

# compile the go binary0
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/redirects

# second - running stage
# we start from scratch
FROM scratch

WORKDIR /

COPY --from=build /bin/redirects /redirects

EXPOSE 8090

USER 1001:1001

CMD ["/redirects"]