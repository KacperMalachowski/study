FROM ubuntu:20.04

ARG DEBIAN_FRONTEND=noninteractive

ENV PYTHONBUFFERED=1

RUN apt-get update -y && apt install -y sudo telnetd vim systemctl && apt-get clean

RUN adduser -gecos --disabled-password --shel /bin/bash telnet

RUN echo "telnet:telnet" | chpasswd

EXPOSE 23

CMD systemctl start inetd; while [ true ]; do sleep 60; done