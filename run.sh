#!/usr/bin/env bash
docker run --env-file ./env.list --name bot -d -v /root/data:/go/data maddevs/nambataxi-telegram-bot