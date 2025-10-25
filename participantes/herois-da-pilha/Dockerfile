FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o /usr/local/bin/app ./main.go

EXPOSE 18020

CMD ["app"]
