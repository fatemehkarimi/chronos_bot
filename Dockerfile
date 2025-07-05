FROM golang:1.24.3 as build

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN make

FROM ubuntu:22.04 as run
WORKDIR /app
COPY --from=build /app/build/chronos_bot ./chronos_bot

EXPOSE 8080
CMD ["./chronos_bot"]