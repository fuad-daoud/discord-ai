FROM fuaddaoud/golang-ffmpeg:1.22.4-bookworm

RUN mkdir /src
RUN mkdir /src/files
RUN mkdir /src/files/wav

ADD discord-ai /src
ADD dca /src

ENV TOKEN="MTI1MzI4NjI0MzY4NTYzMDAyNA.G6aw13.X4Vg9L9BfRWkNJcVgSbcLnqe_GbKoydkhJ9krw"
ENV DEEPGRAM_API_KEY="b3e84a4a52bf9a59b9be90b1fe40af900adaef52"
ENV OPENAI_API_KEY="sk-proj-AsgPdFnfbcgSNTBdZivIT3BlbkFJPVWizOOQqwPygX2ctH78"
ENV RESPEECHER_API_KEY="DgB1A7jQlUBPEbKjH490bg"

WORKDIR /src

EXPOSE 8080
CMD ["./discord-ai"]