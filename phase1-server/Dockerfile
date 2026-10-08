FROM golang:1.22-alpine AS builder
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY  . .
RUN go build -o u-short

FROM alpine:3.19
ENV GOPROXY=https://goproxy.cn,direct
COPY --from=builder /app/u-short .
EXPOSE 8080
CMD ["./u-short"]
