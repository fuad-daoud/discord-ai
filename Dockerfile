FROM alpine:3.20.1

RUN apk add --no-cache ffmpeg go musl-dev

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin


RUN mkdir /src
RUN mkdir /src/files
RUN mkdir /src/files/wav

ADD discord-ai /src
ADD dca /src

ENV TOKEN="MTI1MzI4NjI0MzY4NTYzMDAyNA.G6aw13.X4Vg9L9BfRWkNJcVgSbcLnqe_GbKoydkhJ9krw"
ENV DEEPGRAM_API_KEY="b3e84a4a52bf9a59b9be90b1fe40af900adaef52"
ENV OPENAI_API_KEY="sk-proj-AsgPdFnfbcgSNTBdZivIT3BlbkFJPVWizOOQqwPygX2ctH78"
ENV RESPEECHER_API_KEY="DgB1A7jQlUBPEbKjH490bg"
ENV NEO4J_DATABASE_URL="neo4j://localhost:7687"
ENV NEO4J_DATABASE_USER="neo4j"
ENV NEO4J_DATABASE_PASSWORD="neo4j"



WORKDIR /src

EXPOSE 8080
CMD ["./discord-ai"]