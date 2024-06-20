FROM golang:1.22.4-bookworm

RUN mkdir -p /src/
ADD . /src/

WORKDIR /src
RUN go build -o /src/discord-ai cmd/ai/main/main.go

RUN apt update
RUN apt install /src/ffmpeg.deb -y

ENV TOKEN="MTI1MzI4NjI0MzY4NTYzMDAyNA.G6aw13.X4Vg9L9BfRWkNJcVgSbcLnqe_GbKoydkhJ9krw"
ENV DEEPGRAM_API_KEY="b3e84a4a52bf9a59b9be90b1fe40af900adaef52"
ENV OPENAI_API_KEY="sk-proj-AsgPdFnfbcgSNTBdZivIT3BlbkFJPVWizOOQqwPygX2ctH78"
ENV RESPEECHER_API_KEY="DgB1A7jQlUBPEbKjH490bg"

WORKDIR /src

CMD ["./discord-ai"]