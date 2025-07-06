FROM golang:1.24.3 as build

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN make

FROM ubuntu:22.04 as run
# Install CA certificates to fix SSL certificate verification
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
ENV TZ="Asia/Tehran"
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone


WORKDIR /app
COPY --from=build /app/config.yaml ./config.yaml
COPY --from=build /app/build/chronos_bot ./chronos_bot

EXPOSE 8080
CMD ["./chronos_bot"]