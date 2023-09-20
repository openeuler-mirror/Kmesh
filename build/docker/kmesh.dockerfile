#
# Dockerfile for building openEuler kmesh docker image.
# 
# Usage:
# docker build -f kmesh.dockerfile -t kmesh:0.3.0 .
# docker run -itd --privileged=true -v /mnt:/mnt -v /sys/fs/bpf:/sys/fs/bpf -v /lib/modules:/lib/modules --name kmesh kmesh:0.3.0
#

# base image
FROM openeuler/openeuler:23.09

# container work directory
WORKDIR /kmesh

# copy current directory files to container work directory
# start_kmesh.sh and may also need kmesh rpm package
ADD . /kmesh

# install pkg dependencies 
# RUN yum install -y kmod util-linux kmesh
RUN yum install -y kmod \
    && yum install -y util-linux \
    && yum install -y Kmesh \
    && yum clean all \
    && rm -rf /var/cache/yum

RUN cp /lib/modules/kmesh/kmesh.ko .

# expose port
EXPOSE 6789

# start kmesh service
ENTRYPOINT ["/bin/sh", "./start_kmesh.sh", "-enable-kmesh", "-enable-ads=true"]
