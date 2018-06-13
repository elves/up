FROM golang:1.10-stretch

# Install pandoc
RUN wget -O /tmp/pandoc.deb https://github.com/jgm/pandoc/releases/download/2.2.1/pandoc-2.2.1-1-amd64.deb && \
    dpkg -i /tmp/pandoc.deb && \
    rm /tmp/pandoc.deb

# Install dependencies from APT
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y git make rsync

# Build app
COPY app /app
RUN go build -o /app/up /app/up.go
RUN mkdir /data

CMD ["/app/up", "-secret", "/data/secret", "-hook", "/app/hook"]

EXPOSE 80
