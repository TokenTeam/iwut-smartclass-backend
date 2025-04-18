FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd

FROM alpine:latest

RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY --from=builder /app/server .

ARG PORT
ARG DEBUG
ARG DATABASE
ARG TENCENT_SECRET_ID
ARG TENCENT_SECRET_KEY
ARG BUCKET_URL
ARG OPENAI_ENDPOINT
ARG OPENAI_KEY
ARG OPENAI_MODEL

ENV PORT=${PORT}
ENV DEBUG=${DEBUG}
ENV DATABASE=${DATABASE}
ENV TENCENT_SECRET_ID=${TENCENT_SECRET_ID}
ENV TENCENT_SECRET_KEY=${TENCENT_SECRET_KEY}
ENV BUCKET_URL=${BUCKET_URL}
ENV OPENAI_ENDPOINT=${OPENAI_ENDPOINT}
ENV OPENAI_KEY=${OPENAI_KEY}
ENV OPENAI_MODEL=${OPENAI_MODEL}

EXPOSE 8080

CMD ["./server"]
