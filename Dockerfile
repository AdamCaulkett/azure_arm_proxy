FROM rightscale/ops_os_base

MAINTAINER slava@rightscale.com

WORKDIR /srv/azure_arm_proxy
COPY bin/entrypoint.sh /srv/azure_arm_proxy/entrypoint.sh
COPY binary /srv/azure_arm_proxy/binary
RUN tar zxvf /srv/azure_arm_proxy/binary/azure_v2-linux-amd64.tgz
EXPOSE 8083
CMD ["web"]
ENTRYPOINT ["./entrypoint.sh"]

ARG gitref=unknown
LABEL git.ref=${gitref}