FROM golang:1.22.4-bullseye as build

ADD . /workspace

WORKDIR /workspace

RUN go mod download
RUN go build -o "./bin/run" "./cmd/ai/main/main.go"


FROM debian:bullseye-slim


RUN mkdir -p -v /src/bin
RUN mkdir -p -v /src/integrations/cohere

COPY --from=build /workspace/bin/run /src/bin
COPY --from=build /workspace/bin/ffmpeg /src/bin
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /workspace/integrations/cohere/cohere-inst.txt /src/integrations/cohere

WORKDIR /src

EXPOSE 8080
CMD ["./bin/run"]