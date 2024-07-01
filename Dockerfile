FROM golang:1.22.4-bullseye as build

ADD . /workspace

WORKDIR /workspace

RUN go mod download
RUN go build -o "./discord-ai" "./cmd/ai/main/main.go"


FROM debian:bullseye-slim


RUN mkdir /src
RUN mkdir /src/files
RUN mkdir /src/files/wav

COPY --from=build /workspace/discord-ai /src
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /src

EXPOSE 8080
CMD ["./discord-ai"]