# syntax=docker/dockerfile:1

ARG KIND=plugins
ARG USERNAME=codeclarity
ARG PLUGINNAME=plugin

# DEV IMAGE
FROM golang:1.23.0 AS plugin-dev
ARG KIND
ARG PLUGINNAME

WORKDIR /
RUN wget https://github.com/alexdobin/STAR/releases/download/2.7.11b/STAR_2.7.11b.zip
RUN apt-get update && apt-get install unzip
RUN unzip STAR_2.7.11b.zip
RUN chmod a+x ./STAR_2.7.11b/Linux_x86_64_static/STAR
RUN cp ./STAR_2.7.11b/Linux_x86_64_static/STAR /usr/local/bin/STAR

WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
ADD .cloud/docker/config/* /root

RUN go install github.com/air-verse/air@latest
CMD air -c .air.toml

# DEBUG IMAGE
FROM golang:1.23.0-alpine AS plugin-debug
ARG KIND
ARG PLUGINNAME
ENV KIND ${KIND}
ENV PLUGINNAME ${PLUGINNAME}

WORKDIR /
RUN wget https://www.bioinformatics.babraham.ac.uk/projects/fastqc/fastqc_v0.12.1.zip
RUN unzip fastqc_v0.12.1.zip
RUN rm fastqc_v0.12.1.zip
RUN ln -s /FastQC/fastqc /usr/local/bin/fastqc

WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
ADD .cloud/docker/config/* /root
RUN apk add --no-cache git msttcorefonts-installer fontconfig openjdk21-jre perl
RUN update-ms-fonts

RUN go install github.com/go-delve/delve/cmd/dlv@latest
ADD .cloud/docker/config/* /root
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
CMD /go/bin/dlv debug . --headless --listen=:40000 --accept-multiclient
# CMD tail -f /dev/null


# BUILD IMAGE
FROM golang:1.23.0-alpine AS plugin-build
ARG KIND
ARG PLUGINNAME
COPY . /codeclarity/${KIND}/${PLUGINNAME}
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
ADD .cloud/docker/config/* /root
RUN rm go.work
RUN apk update && apk upgrade --no-cache
RUN apk add git
RUN [ "go", "build", "-o", "plugin" ]

# PRODUCTION IMAGE
FROM alpine:3.20.2 AS plugin
ARG KIND
ARG USERNAME
ARG PLUGINNAME
RUN adduser -D ${USERNAME}
USER ${USERNAME}
WORKDIR /home/${USERNAME}/${PLUGINNAME}
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/${KIND}/${PLUGINNAME}/plugin .
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/${KIND}/${PLUGINNAME}/config.json .

RUN mkdir /home/${USERNAME}/private

ENTRYPOINT [ "./plugin" ]
