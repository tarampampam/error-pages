# Common request to the index page
GET http://{{ host }}:{{ port }}/
X-Code: 410

HTTP 410

[Asserts]
header "Content-Type" contains "text/html"
header "X-Robots-Tag" == "noindex"
body contains "410"
body contains "Gone"

# X-Code with X-Format
GET http://{{ host }}:{{ port }}/
X-Code: 410
X-Format: text/html;q=0.9,application/xhtml+xml;q=0.9,application/json,image/avif,image/webp,*/*;q=0.8

HTTP 410

[Asserts]
header "Content-Type" contains "application/json"
header "X-Robots-Tag" == "noindex"
jsonpath "$.error" == true
jsonpath "$.code" == "410"
jsonpath "$.message" == "Gone"

# Error pages should ignore X-Code
GET http://{{ host }}:{{ port }}/502.html
X-Code: 410

HTTP 200

[Asserts]
header "Content-Type" contains "text/html"
header "X-Robots-Tag" == "noindex"
body contains "502"
body contains "Bad Gateway"
