FROM golang:1.22

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bachtrango

EXPOSE 80

ENTRYPOINT ["/build/bachtrango"]