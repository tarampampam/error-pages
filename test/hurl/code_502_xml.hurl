# The common request
GET http://{{ host }}:{{ port }}/502.html
Content-Type: application/xml;charset=UTF-8
X-Original-URI: foo
X-Namespace: bar
X-Ingress-Name: baz
X-Service-Name: aaa
X-Service-Port: bbb
X-Request-ID: ccc

HTTP/* 200

[Asserts]
header "Content-Type" contains "application/xml"
header "X-Robots-Tag" == "noindex"
xpath "string(//error/code)" == "502"
xpath "string(//error/message)" == "Bad Gateway"
xpath "string(//error/description)" == "The server received an invalid response from the upstream server"
xpath "string(//error/details/originalURI)" == "foo"
xpath "string(//error/details/namespace)" == "bar"
xpath "string(//error/details/ingressName)" == "baz"
xpath "string(//error/details/serviceName)" == "aaa"
xpath "string(//error/details/servicePort)" == "bbb"
xpath "string(//error/details/requestID)" == "ccc"
xpath "string(//error/details/forwardedFor)" == "ddd"
xpath "string(//error/details/host)" == "fff"
xpath "string(//error/details/timestamp)" exists

# X-Format in the action
GET http://{{ host }}:{{ port }}/502.html
X-Format: text/xml

HTTP/* 200

[Asserts]
header "Content-Type" contains "application/xml"
header "X-Robots-Tag" == "noindex"
xpath "string(//error/code)" == "502"
xpath "string(//error/message)" == "Bad Gateway"
xpath "string(//error/description)" == "The server received an invalid response from the upstream server"
