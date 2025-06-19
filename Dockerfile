FROM golang:1.24.1 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
RUN apt-get install -y gcc

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./
RUN go build -ldflags='-s -w' -o /bin/application ./cmd/application
RUN go build -ldflags="-s -w" -o /bin/cli ./cmd/cli


FROM golang:1.24.1-alpine3.21

WORKDIR /app

COPY --from=builder /bin/application /bin/application
COPY --from=builder /bin/cli /bin/cli

CMD ["/bin/application"]
