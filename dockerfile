FROM ghcr.io/linuxserver/baseimage-alpine:3.16

###############################################################################
# YTDL-RSS INSTALL

# COPY root/ /
WORKDIR /config
RUN apk update --no-cache
RUN apk upgrade --no-cache
RUN apk add --update bash
RUN apk --no-cache add ca-certificates python3 py3-pip ffmpeg tzdata nano curl go git make musl-dev
RUN ln -sf python3 /usr/bin/python
RUN python3 -m ensurepip
RUN pip3 install --no-cache --upgrade pip setuptools
RUN python3 -m pip install -U yt-dlp
RUN export GOPATH=/root/go
RUN export PATH=${GOPATH}/bin:/usr/local/go/bin:$PATH
RUN export GOBIN=$GOROOT/bin
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
RUN export GO111MODULE=on
RUN cp /usr/share/zoneinfo/Australia/Melbourne /etc/localtime
RUN echo "Australia/Melbourne" >  /etc/timezone
RUN wget -O /tmp/DownloadYouTubeGo.tar.gz https://github.com/awirthy/DownloadYouTubeGo/archive/refs/tags/v1.15.tar.gz
RUN mkdir -p /opt/DownloadYouTubeGo
RUN tar zxf /tmp/DownloadYouTubeGo.tar.gz -C /opt/DownloadYouTubeGo
RUN echo "#!/bin/sh" >> /etc/periodic/15min/DownloadYouTubeGo
RUN echo "/opt/DownloadYouTubeGo/DownloadYouTubeGo-1.15/DownloadYouTubeGo.sh" >> /etc/periodic/15min/DownloadYouTubeGo
RUN chmod 755 /opt/DownloadYouTubeGo/DownloadYouTubeGo-1.15/DownloadYouTubeGo.sh
RUN chmod 755 /etc/periodic/15min/DownloadYouTubeGo
CMD ["crond", "-f","-l","8"]
    
###############################################################################
# CONTAINER CONFIGS

ENV EDITOR="nano" \
#ENV TZ="Australia/Melbourne" \
