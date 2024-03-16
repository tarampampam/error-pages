GET http://{{ host }}:{{ port }}/not-found

HTTP 404

[Asserts]
header "Content-Type" contains "text/html"
body contains "The server can not find the requested page"
