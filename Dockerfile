FROM golang:1.11.1-stretch

# Config Environment Variable
ENV TINI_VERSION v0.13.0

# # Setup custom GO Environment Variable (comment to use the default)
# ENV GOPATH /go
ENV GOROOT /usr/local/go

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y git wget curl

# Add Python stuff
RUN apt-get install -y python3 python3-dev python3-pip python3-scipy python3-numpy && \
    python3 -m pip install scikit-learn

# Install SVM
WORKDIR "/tmp"
RUN wget http://www.csie.ntu.edu.tw/~cjlin/cgi-bin/libsvm.cgi?+http://www.csie.ntu.edu.tw/~cjlin/libsvm+tar.gz -O libsvm.tar.gz && \
        tar -xvzf libsvm.tar.gz && \
        mv libsvm-*/* ./ && \
        make && \
        cp svm-scale /usr/local/bin/ && \
        cp svm-predict /usr/local/bin/ && \
        cp svm-train /usr/local/bin/ && \
        rm -rf *

# Install mosquitto
RUN apt-get update && apt-get install -y apt-transport-https mosquitto-clients mosquitto

# Install Parsin to GOPATH/src
WORKDIR $GOPATH/src/ParsinServer/
ADD . .
RUN go get ./... && \
    go build && \
    echo "\ninclude_dir /code/mosquitto" >> /etc/mosquitto/mosquitto.conf

# # Setup supervisor
# RUN apt-get update && \
#     apt-get install -y supervisor

# # Add supervisor
# ADD ./configs/supervisord.conf /etc/supervisor/conf.d/

# # Add Tini
# ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
# RUN chmod +x /tini

# ENTRYPOINT ["/tini", "--"]

# # Startup
# CMD ["/usr/bin/supervisord"]

# Expose the ports
Expose 8003/tcp
Expose 1883/tcp

# Run the server
ENTRYPOINT ["/go/src/ParsinServer/ParsinServer"]

