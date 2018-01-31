FROM alpine:latest

COPY ./kubermaster /

EXPOSE 9090
CMD ["./kubermaster"]
