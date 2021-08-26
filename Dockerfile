# Usage:
# ```
# $ docker run -d --rm u1and0/grep-server [options]
# ```

FROM golang:1.17-alpine3.14 AS builder
WORKDIR /go/src/github.com/u1and0/grep-server
COPY main.go .
COPY go.mod .
COPY go.sum .
COPY cmd/ cmd/
RUN go build -o /usr/bin/grep-server

FROM alpine:3.14 AS runner
COPY --from=pandoc/core:latest /usr/local/bin /usr/local/bin
COPY --from=builder /usr/bin /usr/bin
RUN apk add --upgrade --no-cache \
        curl \
        ffmpeg \
        poppler-utils \
        ripgrep \
        lua5.3-libs  # Needed for pandoc
RUN RGA_BINARY=https://github.com/phiresky/ripgrep-all/releases/download/v0.9.6/ripgrep_all-v0.9.6-x86_64-unknown-linux-musl.tar.gz &&\
    curl -L $RGA_BINARY | tar -xvzf- &&\
    cp ripgrep_all*/rga* /usr/bin &&\
    rm -rf ripgrep_all*

EXPOSE 8080
ENTRYPOINT ["/usr/bin/grep-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Running grep-server"\
      version="2.0.0"
