FROM golang:1.17-alpine3.13 AS  builder
COPY . /server
WORKDIR /server
RUN go build -o main main.go 

#run stage
FROM alpine:3.13
WORKDIR /server
COPY --from=builder /server/main .
COPY ./environment .
COPY start.sh .
COPY wait-for.sh .
COPY room.txt .
EXPOSE 8080
CMD ["/server/main"]
ENTRYPOINT [ "/server/start.sh" ]

# FROM server:version1.9
# WORKDIR /server
# # COPY ./environment .
# # COPY room.txt .
# EXPOSE 8080
# CMD ["/server"]
