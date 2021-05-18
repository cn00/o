FROM debian

ADD bin/octo-api /octo/bin/
ADD static/file /static/
EXPOSE 8080
ENV GIN_MODE release

CMD ["/octo/bin/octo-api", "-conf", "/octo/etc/config-api.tml"]
