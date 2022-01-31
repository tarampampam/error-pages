local formats = { 'application/json', 'application/xml', 'text/html', 'text/plain' }

request = function()
    wrk.headers["X-Namespace"] = "NAMESPACE_" .. tostring(math.random(0, 99999999))
    wrk.headers["X-Request-ID"] = "REQ_ID_" .. tostring(math.random(0, 99999999))
    wrk.headers["Content-Type"] = formats[ math.random( 0, #formats - 1 ) ]

    return wrk.format("GET", "/500.html?rnd=" .. tostring(math.random(0, 99999999)), nil, nil)
end
