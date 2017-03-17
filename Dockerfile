# Containter for Namba Taxi Telegram Bot
FROM golang:1.8
MAINTAINER Oleg Puzanov <puzanov@gmail.com>
RUN apt-get update -y && apt-get install -y
RUN go get github.com/maddevsio/nambataxi-telegram-bot
RUN go build github.com/maddevsio/nambataxi-telegram-bot
COPY config.yaml /go/config.yaml
RUN ./nambataxi-telegram-bot