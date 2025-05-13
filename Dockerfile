FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd

FROM alpine:latest

RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY --from=builder /app/server .

ENV PORT=""
ENV DEBUG=""
ENV DATABASE=""
ENV TENCENT_SECRET_ID=""
ENV TENCENT_SECRET_KEY=""
ENV BUCKET_URL=""
ENV OPENAI_ENDPOINT=""
ENV OPENAI_KEY=""
ENV OPENAI_MODEL=""
ENV INFO_SIMPLE=""
ENV GET_WEEK_SCHEDULES=""
ENV SEARCH_LIVE_COURSE_LIST=""

EXPOSE 8080

CMD ["./server"]
