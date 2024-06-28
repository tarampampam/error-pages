# Templates

Creating templates is a straightforward process, even for those unfamiliar
with [Go Template](https://pkg.go.dev/text/template). Here are a few things to keep in mind:

- The template should be a single page, without additional `css` or `js` files. However, you can load them from a
  CDN or other GitHub repositories using [jsdelivr.com](https://www.jsdelivr.com/)
- Be sure to include the `<meta name="robots" content="nofollow,noarchive,noindex">` tag in the header
- You can use special "placeholders" (wrapped in `{{` and `}}`) for the rendering error code, message, and other
  details

Built-in "placeholders" and functions with their examples can be found in the following files:

- [props.go](../internal/template/props.go)
- [template.go](../internal/template/template.go)
