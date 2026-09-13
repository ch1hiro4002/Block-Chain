FROM scratch
WORKDIR /app
COPY bin/Block-Chain ./blockchain
EXPOSE 3000 4000 5000 9997 9998 9999
ENTRYPOINT ["/app/blockchain"]
