FROM alpine:3.5

ADD ./pv-cleaner-operator /pv-cleaner-operator

ENTRYPOINT ["/pv-cleaner-operator"]