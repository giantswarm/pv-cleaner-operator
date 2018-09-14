FROM alpine:3.8

ADD ./pv-cleaner-operator /pv-cleaner-operator

ENTRYPOINT ["/pv-cleaner-operator"]
