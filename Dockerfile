# Usage:
# ```
# $ docker run -d --rm u1and0/grep-server [options]
# ```

# BUILD IMAGE
FROM golang:1.14.3-buster AS builder
# Install ripgrep-all
WORKDIR /tmp
RUN curl -fsSL https://github.com/phiresky/ripgrep-all/releases/download/v0.9.6/ripgrep_all-v0.9.6-x86_64-unknown-linux-musl.tar.gz | tar -xzv -C .
# Install dependencies package
RUN apt update &&\
    apt install -y ripgrep \
                pandoc \
                poppler-utils \
                ffmpeg \
                cargo
# Build grep-server
COPY ./main.go /go/src/github.com/u1and0/grep-server/main.go
WORKDIR /go/src/github.com/u1and0/grep-server
# For go module using go-pipeline
# ENV GO111MODULE=on
# RUN apk --update --no-cache add git &&\
RUN go build -o /go/bin/grep-server

# RUN IMAGE
FROM debian:buster as runnner
WORKDIR /
COPY --from=builder /tmp/ripgrep_all-v0.9.6-x86_64-unknown-linux-musl/rga /usr/bin/rga
COPY --from=builder /tmp/ripgrep_all-v0.9.6-x86_64-unknown-linux-musl/rga-preproc /usr/bin/rga-preproc
COPY --from=builder /usr/bin/rg /usr/bin/rg
COPY --from=builder /usr/bin/pandoc /usr/bin/pandoc
COPY --from=builder /usr/bin/pdftotext /usr/bin/pdftotext
COPY --from=builder /usr/bin/ffmpeg  /usr/bin/ffmpeg
COPY --from=builder /go/bin/grep-server /usr/bin/grep-server
EXPOSE 8080
ENTRYPOINT ["/usr/bin/grep-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Running grep-server"\
      version="v0.0.0"
