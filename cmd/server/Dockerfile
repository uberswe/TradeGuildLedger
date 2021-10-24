FROM ubuntu:focal

WORKDIR /app
# Prevent prompts
ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update &&\
    apt-get install --assume-yes ca-certificates libx11-dev apt-utils gcc make gcc-mingw-w64 nodejs npm &&\
    apt-get install --assume-yes golang-1.16

RUN export PATH=/usr/lib/go-1.16/bin:$PATH

RUN ln -s /usr/lib/go-1.16 /usr/local/go

RUN export PATH=/usr/lib/go/bin:$PATH

# This needs to be built with the root project directory as context
COPY . .

RUN /usr/local/go/bin/go build -o /server cmd/server/main.go

RUN npm install

ENV HTTP_PORT=:80

EXPOSE 80

CMD [ "/server" ]