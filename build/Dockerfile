FROM registry.access.redhat.com/ubi7/ubi-minimal:7.6

ENV OPERATOR=/usr/local/bin/useless-operator \
    USER_UID=1001 \
    USER_NAME=useless-operator

# install operator binary
COPY build/_output/bin/useless-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
