# Usage:
# ```
# $ docker run -d --rm u1and0/grep-server [options]
# ```

# BUILD IMAGE
FROM u1and0/archlinux
# FROM u1and0/archlinux AS builder
USER root
COPY build/mirrorlist /etc/pacman.d/mirrorlist
# Build grep-server
RUN sudo -u u1and0 yay -Syu --noconfirm --afterclean --removemake --save \
        ripgrep-all \
        go
    # pacman -Qtdq | xargs -r pacman --noconfirm -Rcns &&\
    # : "Remove caches forcely" &&\
    # : "[error] yes | pacman -Scc" &&\
    # rm -rf /home/u1and0/.cache &&\
WORKDIR /go/src/github.com/u1and0/grep-server
COPY main.go .
COPY go.mod .
COPY cmd/ cmd/
RUN go build -o /usr/bin/grep-server
    # pacman --noconfirm -Rcns go &&\
    # pacman -Scc --noconfirm

# FROM archlinux/base as runnner
# COPY --from=builder /usr/bin /usr/bin

EXPOSE 8080
ENTRYPOINT ["/usr/bin/grep-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Running grep-server"\
      version="v1.0.0"
