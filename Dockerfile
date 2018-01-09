FROM alpine:3.7

ADD ./pv-cleaner-operator /pv-cleaner-operator

ENTRYPOINT ["/pv-cleaner-operator"]