# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM gocv/opencv:latest AS base
#FROM golang:1.20-bullseye
WORKDIR /app

RUN mkdir -p ./html
COPY ./html/*.html ./html

RUN mkdir -p ./server
COPY /server/*.go ./server

RUN mkdir -p ./test_data
COPY /test_data/* ./test_data

COPY *.go ./
COPY go.mod go.sum ./

RUN go mod download

RUN CGO_ENABLED=1 GOOS=linux go build -o /goremotecontrol_web

EXPOSE 8080

CMD [ "/goremotecontrol_web"]