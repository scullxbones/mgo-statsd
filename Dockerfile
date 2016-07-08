FROM golang

ADD . /usr/src/mgo-statsd
WORKDIR /usr/src/mgo-statsd

RUN ./build.sh

ENTRYPOINT [ "./mgo-statsd" ]
CMD [ "--help" ]
