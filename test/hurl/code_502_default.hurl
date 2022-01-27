GET http://{{ host }}:{{ port }}/502.html

HTTP/* 200

[Asserts]
header "Content-Type" contains "text/html"
header "X-Robots-Tag" == "noindex"
body contains "502"
body contains "Bad Gateway"
body contains "The server received an invalid response from the upstream server"
