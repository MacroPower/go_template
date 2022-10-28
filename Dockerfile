FROM golang:1.19

COPY go_template /usr/local/bin/

ENTRYPOINT ["go_template"]
