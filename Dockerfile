FROM rightscale/ops_os_base

MAINTAINER slava@rightscale.com

RUN mkdir -p /root/binary
WORKDIR /root/binary
COPY binary /root/binary
RUN tar zxvf /root/binary/azure_v2-linux-amd64.tgz
RUN ls /root/binary/
RUN /root/binary/azure_v2/azure_v2 --listen="localhost:8083" --prefix="/azure_v2"

EXPOSE 8083