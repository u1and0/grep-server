# Usage:
# ```
# $ docker run -d --rm u1and0/grep-server [options]
# ```

FROM pandoc/core AS builder
RUN apk add --upgrade --no-cache \
        curl \
        ffmpeg \
        poppler-utils \
        ripgrep \
        go
RUN RGA_BINARY=https://github.com/phiresky/ripgrep-all/releases/download/v0.9.6/ripgrep_all-v0.9.6-x86_64-unknown-linux-musl.tar.gz &&\
    curl -LO $RGA_BINARY &&\
    tar -xvf "$(basename $RGA_BINARY)" &&\
    cp ripgrep_all*/rga* /usr/bin
# RUN apt update && apt install -y build-essential pandoc poppler-utils ffmpeg ripgrep
#
# RUN apt update && apt install -y cargo
# RUN cargo install ripgrep_all

# RUN apt update && apt install -y go
WORKDIR /go/src/github.com/u1and0/grep-server
COPY main.go .
COPY go.mod .
COPY go.sum .
COPY cmd/ cmd/
RUN go build -o /usr/bin/grep-server

FROM alpine as runner
COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /usr/bin /usr/bin
COPY --from=builder /lib /lib
# Why need it `apk upgrade`? but doesn't work unless do it.
RUN apk update && apk add --upgrade ripgrep

EXPOSE 8080
ENTRYPOINT ["/usr/bin/grep-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Running grep-server"\
      version="2.0.0"
