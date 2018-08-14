FROM centos:7

WORKDIR /opt/greeter

VOLUME ["/opt/greeter/conf/"]

RUN mkdir -p /opt/greeter/
RUN mkdir -p /opt/greeter/conf

ADD xlsoa.example.greeter /opt/greeter/

ENTRYPOINT ["/opt/greeter/xlsoa.example.greeter"]
