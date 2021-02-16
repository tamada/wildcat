FROM alpine:3.10.1
LABEL maintainer="Haruai Tamada" \
      wildcat-version="1.0.0" \
      description="another implementation of wc (word count)"

RUN    adduser -D wildcat \
    && apk --no-cache add curl tar \
    && curl -s -L -O https://github.com/tamada/wildcat/releases/download/v1.0.0/wildcat-1.0.0_linux_amd64.tar.gz \
#    && curl -s -L -o wildcat-1.0.0_linux_amd64.tar.gz https://www.dropbox.com/s/9av696fxcoz92o4/wildcat-1.0.0_linux_amd64.tar.gz?dl=0 \
    && tar xfz wildcat-1.0.0_linux_amd64.tar.gz          \
    && mv wildcat-1.0.0 /opt                             \
    && ln -s /opt/wildcat-1.0.0 /opt/wildcat             \
    && ln -s /opt/wildcat/wildcat /usr/local/bin/wildcat \
    && rm wildcat-1.0.0_linux_amd64.tar.gz

ENV HOME="/home/wildcat"

WORKDIR /home/wildcat
USER    wildcat

ENTRYPOINT [ "wildcat" ]
