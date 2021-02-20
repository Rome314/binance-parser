FROM golang:1.14 AS builder
WORKDIR $GOPATH/src/github.com/rome314/binance-parser
COPY go.mod ./
RUN go mod download
COPY . ./
RUN make test
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app cmd/*.go

FROM scratch
COPY --from=builder /app ./
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT ["./app"]