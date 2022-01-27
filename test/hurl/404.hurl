GET http://{{ host }}:{{ port }}/not-found

HTTP/* 404

[Asserts]
header "Content-Type" contains "text/plain"
body contains "Wrong request URL"
