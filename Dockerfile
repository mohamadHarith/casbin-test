# Build Stage
FROM golang:1.10 AS build-stage

LABEL app="build-casbin-test"
LABEL REPO="https://github.com/ljdursi/casbin-test"

ENV GOPATH=/gopath \
    PROJPATH=/gopath/src/github.com/ljdursi/casbin-test

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOPATH/bin

RUN mkdir -p $GOPATH/bin

ADD . /gopath/src/github.com/ljdursi/casbin-test
WORKDIR /gopath/src/github.com/ljdursi/casbin-test

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN make build-alpine

# Final Stage
FROM golang:1.10

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/ljdursi/casbin-test"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/casbin-test/bin

WORKDIR /opt/casbin-test/bin

COPY --from=build-stage /gopath/src/github.com/ljdursi/casbin-test/bin/casbin-test /opt/casbin-test/bin/
COPY model.conf /opt/casbin-test/bin/
COPY policy.csv /opt/casbin-test/bin/
RUN chmod +x /opt/casbin-test/bin/casbin-test

CMD ./casbin-test
