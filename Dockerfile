# Usage:
# ```
# $ docker run -d --rm u1and0/grep-server [options]
# ```

FROM u1and0/archlinux AS builder
USER root
COPY build/mirrorlist /etc/pacman.d/mirrorlist
RUN sudo -u u1and0 yay -Syu --noconfirm --afterclean --removemake --save \
        ripgrep-all \
        go
WORKDIR /go/src/github.com/u1and0/grep-server
COPY main.go .
COPY go.mod .
COPY go.sum .
COPY cmd/ cmd/
RUN go build -o /usr/bin/grep-server

FROM archlinux as runnner
COPY --from=builder /usr/bin /usr/bin
COPY --from=builder /lib /lib

EXPOSE 8080
ENTRYPOINT ["/usr/bin/grep-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Running grep-server"
