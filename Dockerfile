# Containter for Namba Taxi Telegram Bot
FROM golang:1.8
LABEL Description="Order a Namba Taxi cab via Telegram" Vendor="Mad Devs" Version="1.4"
MAINTAINER Oleg Puzanov <puzanov@gmail.com>
RUN go get -v github.com/maddevsio/nambataxi-telegram-bot
RUN go build -v github.com/maddevsio/nambataxi-telegram-bot
COPY config.production.yaml /go/config.yaml
CMD ./nambataxi-telegram-bot