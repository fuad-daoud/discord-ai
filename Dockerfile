FROM golang:1.22.4-bullseye as build

ADD . /workspace

WORKDIR /workspace

RUN go mod download
RUN go build -o "./discord-ai" "./cmd/ai/main/main.go"


FROM debian:bullseye-slim


RUN mkdir -v /src
RUN mkdir -p -v /src/integrations/cohere

COPY --from=build /workspace/discord-ai /src
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /workspace/integrations/cohere/cohere-inst.txt /src/integrations/cohere

WORKDIR /src

EXPOSE 8080
CMD ["./discord-ai"]