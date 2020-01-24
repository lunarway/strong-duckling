FROM golang:1.13.6 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o strong-duckling .

RUN ./strong-duckling

# FROM scratch

# WORKDIR /app
# COPY main .

# RUN ./main
