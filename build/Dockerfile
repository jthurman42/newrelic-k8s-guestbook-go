# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
FROM golang:1.12.0
RUN go get -u -v github.com/kardianos/govendor

RUN mkdir -p /go/src/app
ADD . /go/src/app

WORKDIR /go/src/app
RUN CGO_ENABLED=0 GOOS=linux make compile-only

FROM scratch
WORKDIR /app
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /go/src/app/bin/ .
COPY assets/ .
ENTRYPOINT ["/app/guestbook"]
CMD ["guestbook"]
EXPOSE 3000
