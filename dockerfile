FROM gcr.io/distroless/static
LABEL Brayden Winterton <brayden_winterton@byu.edu>

ARG NAME

COPY ${NAME} /server

ENTRYPOINT ["/server"]
