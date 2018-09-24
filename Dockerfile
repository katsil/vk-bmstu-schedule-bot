FROM ubuntu:18.04
LABEL author='George Gabolaev'

ENV GOVERSION=1.9.6
ENV REPO=github.com/BMSTU-bots/vk-bmstu-schedule-bot

# Basic tools and configurations
RUN apt update
RUN apt install -y git vim wget curl locales
RUN locale-gen en_US.UTF-8
ENV LC_ALL=en_US.UTF-8

# # go installation
RUN wget https://storage.googleapis.com/golang/go$GOVERSION.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go$GOVERSION.linux-amd64.tar.gz && \
    mkdir go && mkdir go/src && mkdir go/bin && mkdir go/pkg
ENV GOROOT /usr/local/go
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH
RUN mkdir -p "$GOPATH/bin" "$GOPATH/src"
RUN go get -u github.com/golang/dep/cmd/dep

# python installations
RUN apt install -y python3 python3-pip

# parser installation
RUN pip3 install bmstu-schedule

# bot moving
ADD ./ $GOPATH/src/$REPO
WORKDIR $GOPATH/src/$REPO
RUN dep ensure
RUN go build