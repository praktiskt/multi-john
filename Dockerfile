FROM golang
COPY . /go/src
WORKDIR /go/src
RUN CGO_ENABLED=0 go build -o multijohn

FROM phocean/john_the_ripper_jumbo  
WORKDIR /go/src
COPY --from=0 /go/src .
CMD ["./multijohn"]  
