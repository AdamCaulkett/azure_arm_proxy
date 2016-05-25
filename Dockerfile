FROM rightscale/ops_os_base

MAINTAINER slava@rightscale.com
ENV LOG_TYPE=stdout

RUN mkdir -p /root/binary
WORKDIR /root/binary
COPY binary /root/binary
RUN tar zxvf /root/binary/azure_v2-linux-amd64.tgz
RUN ls /root/binary/

EXPOSE 8083

ARG gitref=unknown
LABEL git.ref=${gitref}