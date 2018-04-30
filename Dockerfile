# Build Stage
FROM lacion/docker-alpine:gobuildimage AS build-stage

LABEL app="build-casbin-test"
LABEL REPO="https://github.com/ljdursi/casbin-test"

ENV GOROOT=/usr/lib/go \
    GOPATH=/gopath \
    GOBIN=/gopath/bin \
    PROJPATH=/gopath/src/github.com/ljdursi/casbin-test

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /gopath/src/github.com/ljdursi/casbin-test
WORKDIR /gopath/src/github.com/ljdursi/casbin-test

RUN make build-alpine

# Final Stage
FROM lacion/docker-alpine:latest

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/ljdursi/casbin-test"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/casbin-test/bin

WORKDIR /opt/casbin-test/bin

COPY --from=build-stage /gopath/src/github.com/ljdursi/casbin-test/bin/casbin-test /opt/casbin-test/bin/
RUN chmod +x /opt/casbin-test/bin/casbin-test

CMD /opt/casbin-test/bin/casbin-test