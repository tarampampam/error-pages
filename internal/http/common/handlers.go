package common

import (
	"strings"

	"github.com/valyala/fasthttp"
)

const internalErrorPattern = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="robots" content="noindex, nofollow" />
    <title>Internal error occurred</title>
    <style>
        html,body {background-color: #0e0e0e;color:#fff;font-family:'Nunito',sans-serif;height:100%;margin:0}
        .message {height:100%;align-items:center;display:flex;justify-content:center;position:relative;font-size:1.4em}
        img {padding-right: .4em}
    </style>
</head>
<body>
<div class="message">
    <img src="https://hsto.org/webt/fs/sx/gt/fssxgtssfg689qxboqvjil5yz8g.png" alt="logo" height="32">
    <p>{{ message }}</p>
</div>
</body>
</html>`

func HandleInternalHTTPError(ctx *fasthttp.RequestCtx, statusCode int, message string) {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("text/html; charset=UTF-8")

	_, _ = ctx.WriteString(strings.ReplaceAll(internalErrorPattern, "{{ message }}", message))
}
