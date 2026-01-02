FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache upx

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server ./cmd && upx -9 server

FROM gcr.io/distroless/static-debian13

WORKDIR /app

COPY --from=mwader/static-ffmpeg /ffmpeg /usr/local/bin/

COPY --from=builder /app/server .

ENV TZ="Asia/Shanghai" \
    PORT="" \
    DEBUG="" \
    DATABASE="" \
    LOG_SAVE="" \
    SUMMARY_WORKER_COUNT="" \
    SUMMARY_QUEUE_SIZE="" \
    TENCENT_SECRET_ID="" \
    TENCENT_SECRET_KEY="" \
    BUCKET_URL="" \
    OPENAI_ENDPOINT="" \
    OPENAI_KEY="" \
    OPENAI_MODEL="" \
    INFO_SIMPLE="" \
    GET_WEEK_SCHEDULES="" \
    SEARCH_LIVE_COURSE_LIST=""

EXPOSE 8080

VOLUME ["/app/data"]

CMD ["./server"]
