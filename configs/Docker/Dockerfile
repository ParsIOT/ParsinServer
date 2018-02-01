# sudo docker build -t finddocker .
# sudo docker run -it -p 18003:8003 -p 11883:1883 -v /path/to/host/data/folder:/data finddocker bash
FROM ubuntu:16.04

# Get basics
RUN apt-get update
RUN apt-get -y upgrade
RUN apt-get install -y golang git wget curl vim
#RUN apt-get install -y git wget curl vim
# RUN mkdir /usr/local/work
# ENV GOPATH /usr/local/work
WORKDIR "/tmp"
RUN wget "https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz"
RUN tar -xvf go1.8.3.linux-amd64.tar.gz -C /usr/local/
RUN chown -R root:root /usr/local/go/
ENV GOPATH /root/project
ENV GOROOT /usr/local/go

# Add Python stuff
RUN apt-get install -y python3 python3-dev python3-pip
RUN apt-get install -y python3-scipy python3-numpy
RUN python3 -m pip install scikit-learn

# Install SVM
WORKDIR "/tmp"
RUN wget http://www.csie.ntu.edu.tw/~cjlin/cgi-bin/libsvm.cgi?+http://www.csie.ntu.edu.tw/~cjlin/libsvm+tar.gz -O libsvm.tar.gz
RUN tar -xvzf libsvm.tar.gz
RUN mv libsvm-*/* ./
RUN make
RUN cp svm-scale /usr/local/bin/
RUN cp svm-predict /usr/local/bin/
RUN cp svm-train /usr/local/bin/
RUN rm -rf *

# Install mosquitto
RUN apt-get update && apt-get install -y apt-transport-https
RUN apt-get update
RUN apt-get install -y mosquitto-clients mosquitto

# Install FIND
# WORKDIR "/root"
# RUN go get git@gitlab.com:ParsIOT/ParsinServer.git
# RUN git clone git@gitlab.com:ParsIOT/ParsinServer.git
# WORKDIR "/root/ParsinServer"
WORKDIR "/root/project/"
ADD . /root/project/
#RUN ls -Rlh
#RUN echo $GOPATH
#RUN echo $GOROOT
#RUN which go
WORKDIR "/root/project/src/ParsinServer"
RUN go get ./...
RUN go build
RUN echo "\ninclude_dir /root/find/mosquitto" >> /etc/mosquitto/mosquitto.conf

# Old entrypoint
# ENTRYPOINT git pull && go build && mosquitto -c /root/find/mosquitto/conf -d && ./find -mqtt localhost:1883 -mqttadmin admin -mqttadminpass 123 -mosquitto `pgrep mosquitto` -data /data > log & bash

# Setup supervisor
RUN apt-get update
RUN apt-get install -y supervisor

# Add supervisor
#COPY /root/project/src/ParsinServer/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
ADD ./src/ParsinServer/supervisord.conf /etc/supervisor/conf.d/
# Add Tini
ENV TINI_VERSION v0.13.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]

# Startup
CMD ["/usr/bin/supervisord"]

