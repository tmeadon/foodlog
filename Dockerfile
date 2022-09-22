FROM golang:1.18-alpine

WORKDIR /app

RUN apk add build-base sqlite

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd/
COPY pkg ./pkg/
COPY web ./web/
RUN go build -o foodlog cmd/app/main.go

RUN mkdir -p ./db/sqlite
COPY db/migrations ./db/migrations
 
ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080

CMD [ "./foodlog" ]
