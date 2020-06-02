# Usage:
# ```
# $ docker run -d --rm u1and0/grep-server [options]
# ```

# BUILD IMAGE
FROM u1and0/archlinux AS builder
COPY mirrorlist /etc/pacman.d/mirrorlist
RUN yay -Syu --noconfirm ripgrep-all go
# Build grep-server
USER root
COPY main.go /go/src/github.com/u1and0/grep-server/main.go
WORKDIR /go/src/github.com/u1and0/grep-server
# For go module using go-pipeline
# ENV GO111MODULE=on
# RUN apk --update --no-cache add git &&\
RUN go build -o /go/bin/grep-server

# RUN IMAGE
FROM archlinux/base as runnner
COPY --from=builder /usr/bin/rga /usr/bin/rga
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
