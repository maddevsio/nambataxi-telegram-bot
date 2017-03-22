#!/usr/bin/env bash
docker run --env-file ./env.list --name bot -v /root/data:/go/data maddevs/nambataxi-telegram-bot