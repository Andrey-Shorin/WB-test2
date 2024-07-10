FROM golang
COPY . .
RUN go build -o main .
CMD ["./main"]