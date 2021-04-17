# ftx-lending-bot
***FTX automatically follow the next hour lending rate bot***

Please copy the `.env.example` to `.env` file and set the config data 

> **SUB_ACCOUNT** your ftx sub account name

> **CURRENCY** The currency you want to lend

> **API_KEY** your ftx api key

> **SECRET_KEY** your ftx secret key


FROM golang:1.14-alpine
WORKDIR /ftx-lending-bot
ADD . /ftx-lending-bot
COPY .env .
RUN cd /ftx-lending-bot \
    && go build
ENTRYPOINT ["sh","-c","./FtxLendingBot"]
