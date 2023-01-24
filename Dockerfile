FROM golang:1.19.1-bullseye as builder
RUN apt update && apt install -y upx

WORKDIR /
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -v -ldflags="-s -w"
RUN upx --best --lzma grpc_client; exit 0

FROM scratch as final
COPY --from=builder /grpc_client .

CMD ["./grpc_client"]
