# Vercel Proxy

Simple HTTP proxy designed to run as a Go server on Vercel.

## Deploy

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2FTBXark%2Fvercel-proxy)

This project uses Vercel's Go framework preset and the root `main.go` entrypoint. In production Vercel provides a dynamic `PORT`; the server reads `PORT` and falls back to `3000` for local development.

## Usage

Just add your deployment URL before the URL you want to proxy:

```javascript
fetch("https://project-name.vercel.app/https://example.com?param1=value1&param2=value2")
  .then((res) => res.text())
  .then(console.log.bind(console))
  .catch(console.error.bind(console));
```

```bash
curl -L "https://project-name.vercel.app/https://example.com?param1=value1&param2=value2"
```

The proxy also accepts the single-slash form used by older examples:

```bash
curl -L "https://project-name.vercel.app/https:/example.com?param1=value1&param2=value2"
```

## Local development

Run with the default local port:

```bash
go run main.go
```

Run with an explicit port, matching the way Vercel injects the internal listener port:

```bash
PORT=41857 go run main.go
```

## License

**vercel-proxy** is released under the MIT license. [See LICENSE](LICENSE) for details.
