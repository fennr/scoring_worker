# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
COPY . .
RUN go mod download
RUN go build -o scoring_worker .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/scoring_worker /app/scoring_worker
CMD ["/app/scoring_worker"] 