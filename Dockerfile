FROM alpine:3.20 as build

RUN apk add --no-cache go

ADD . /workspace

WORKDIR /workspace

RUN go mod download
RUN go build -o "./discord-ai" "./cmd/ai/main/main.go"


FROM alpine:3.20


RUN apk add --no-cache ffmpeg

RUN mkdir /src
RUN mkdir /src/files
RUN mkdir /src/files/wav

COPY --from=build /workspace/discord-ai /src
COPY --from=build /workspace/dca /src

WORKDIR /src

EXPOSE 8080
CMD ["./discord-ai"]




#question if I have an `opus` audio file with `ogg` extention can I just stream it to discord voice channel ?
#
#currently I am converting an `mp3` file to `dca` and then stream it.
#
#since I have the option to do `opus` encoding I thought why not just stream it into the voice channel ?
#
#but I have tried something like that in the past but I was not lucky