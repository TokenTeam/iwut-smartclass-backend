# Use the official Golang image as the base image
FROM golang:1.24-alpine

# Install ffmpeg
RUN apk add --no-cache ffmpeg

# Set environment variables
ENV PORT=${PORT}
ENV DEBUG=${DEBUG}
ENV DATABASE=${DATABASE}
ENV TENCENT_SECRET_ID=${TENCENT_SECRET_ID}
ENV TENCENT_SECRET_KEY=${TENCENT_SECRET_KEY}
ENV BUCKET_URL=${BUCKET_URL}
ENV OPENAI_ENDPOINT=${OPENAI_ENDPOINT}
ENV OPENAI_KEY=${OPENAI_KEY}
ENV OPENAI_MODEL=${OPENAI_MODEL}
ENV PROMPT=${PROMPT}

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the pre-built Go app
COPY server .

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
