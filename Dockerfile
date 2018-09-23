FROM golang:latest

MAINTAINER 16yuki0702

# install required packages
RUN apt-get update && apt-get install -y \
  unzip \
  apt-utils 

# install chrome driver
RUN CHROME_DRIVER_VERSION=`curl -sS chromedriver.storage.googleapis.com/LATEST_RELEASE` && \
  wget -N http://chromedriver.storage.googleapis.com/$CHROME_DRIVER_VERSION/chromedriver_linux64.zip && \
  unzip chromedriver_linux64.zip && \
  rm chromedriver_linux64.zip && \
  chown root:root chromedriver && \
  chmod 755 chromedriver && \
  mv chromedriver /usr/bin/chromedriver && \
  sh -c 'wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -' && \
  sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list' && \
  apt-get update && apt-get install -y google-chrome-stable

# install application
ADD . /go/src/github.com/16yuki0702/webanalyzer
ADD template /go/bin/template
RUN go get github.com/PuerkitoBio/goquery \
  github.com/sclevine/agouti \
  github.com/pkg/errors \
  golang.org/x/net/websocket
RUN go install github.com/16yuki0702/webanalyzer

EXPOSE 8080

ENTRYPOINT /go/bin/webanalyzer
