FROM golang:1.14-alpine
ENV Asia/Taipei
WORKDIR /ftx-lending-bot
ADD . /ftx-lending-bot
COPY .env .
RUN cd /ftx-lending-bot \
    && go build
ENTRYPOINT ["sh","-c","./FtxLendingBot"]
