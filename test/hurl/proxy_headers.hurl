GET http://{{ host }}:{{ port }}/502.html
X-Foo: foo
bar: BAR
Baz_blah: baz Baz
NonEx: skip

HTTP 200

[Asserts]
header "X-Foo" == "foo"
header "Bar" == "BAR"
header "Baz_blah" == "baz Baz"
header "NonEx" not exists
