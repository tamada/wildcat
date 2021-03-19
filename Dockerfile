FROM alpine:3.10.1

ARG version=1.1.0

LABEL maintainer="Haruai Tamada" \
      description="another implementation of wc (word count)"

RUN    adduser -D wildcat \
    && apk --no-cache add --update --virtual .builddeps curl tar \
    && curl -s -L -O https://github.com/tamada/wildcat/releases/download/v${version}/wildcat-${version}_linux_amd64.tar.gz \
#    && curl -s -L -o wildcat-${version}_linux_amd64.tar.gz https://www.dropbox.com/s/f60483rucruauiz/wildcat-1.1.0_linux_amd64.tar.gz?dl=0 \
    && tar xfz wildcat-${version}_linux_amd64.tar.gz     \
    && mv wildcat-${version} /opt                        \
    && ln -s /opt/wildcat-${version} /opt/wildcat        \
    && ln -s /opt/wildcat/wildcat /usr/local/bin/wildcat \
    && rm wildcat-${version}_linux_amd64.tar.gz          \
    && apk del --purge .builddeps

ENV HOME="/home/wildcat"

WORKDIR /home/wildcat
USER    wildcat

ENTRYPOINT [ "/opt/wildcat/wildcat" ]
