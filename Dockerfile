## Build
FROM golang:1.20-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /restic-exporter

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=restic/restic /usr/bin/restic /usr/bin

COPY --from=build /restic-exporter /restic-exporter

EXPOSE 9150

ENTRYPOINT ["/restic-exporter"]
