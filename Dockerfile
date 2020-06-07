# Usage:
# ```
# $ docker run -d --rm u1and0/grep-server [options]
# ```

# BUILD IMAGE
FROM u1and0/archlinux
COPY mirrorlist /etc/pacman.d/mirrorlist
# Build grep-server
USER root
COPY main.go /go/src/github.com/u1and0/grep-server/main.go
WORKDIR /go/src/github.com/u1and0/grep-server
RUN sudo -u u1and0 yay -Syu --noconfirm --afterclean --removemake --save \
        ripgrep-all \
        go &&\
    pacman -Qtdq | xargs -r pacman --noconfirm -Rcns &&\
    : "Remove caches forcely" &&\
    : "[error] yes | pacman -Scc" &&\
    rm -rf /home/u1and0/.cache &&\
    go build -o /usr/bin/grep-server &&\
    pacman --noconfirm -Rcns go &&\
    pacman -Scc --noconfirm

EXPOSE 8080
ENTRYPOINT ["/usr/bin/grep-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Running grep-server"\
      version="v0.1.0"
