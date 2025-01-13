# GoMarks

Yet another Golinks manager.

Create smart bookmarks and make GoMarks your default search engine to speed up your browsing experience.

Similar to Duckduckgo <code>!bangs</code> or the multitude of self-hosted Golinks manager (see below for some alternatives).

Bangs are pretty cool but it means requests always transit through Duckduckgo.

## Features

- can act as your default search engine, enabled in 2 easy steps (tested in Firefox and Chrome)
- smart links to redirect to websites
- smart links with placeholder <code>%s</code> to redirect to websites with search engines (example: Amazon)
- if your query doesn't match any smart link, your query is sent to your fallback search engine
- configurable fallback search engine (Google, Duckduckgo, Bing)
- configurable custom fallback search engine (self-hosted Searx, Whoogle, etc.)
- usage statistics
- reset statistics per smart link or for all links
- all information are stored in a sqlite database, allowing to manipulate and backup your data easily
- can run with Docker, Podman or Kubernetes or as a standalone binary

## But Why?

I wanted to learn the basics of Go and create something useful for myself.

Oh, you wanted to know why the logo is a bunny? Historically one of the first Golinks manager was called bunny1. But mostly, I have 3 pet rabbits!

## Docker Installation 

```bash
docker run -d --name gomarks --restart unless-stopped -v /opt/docker/gomarks:/data -p 8080:8080 ghcr.io/sebw/gomarks:latest
```

## Security

There's no authentication, user or certificate management. You MUST secure GoMarks behind something like Let's Encrypt, Authentik, Authelia or Cloudflare.

You can follow [this guide](https://blog.wains.be/2023/2023-01-07-cloudflare-zero-trust-authentik/) to secure GoMarks (SSO + HTTPS) behind Cloudflare and Authentik.

## Screenshots

TODO

## Known Issues

When trying to install GoMarks as the default search engine and the GoMarks instance is hosted behind Cloudflare, I get "Firefox could not install the search engine from: https://gomarks.example.org/opensearch.xml".

Installing from a locally running instance (http://localhost:8080) installs just fine.

I recommend to use this [Firefox add-on](https://addons.mozilla.org/en-GB/firefox/addon/add-custom-search-engine/) if you face the same issue.

## Contributing

GoMarks is created for my needs and I use it a million times a day.

If you feel something is missing and might benefit everyone, please fork and make a pull request. 

If the feature is actually useful for myself, I might implement it in the main branch.

## Alternatives

- [Prologic's golinks](https://git.mills.io/prologic/golinks): the solution I've used for 3-ish years but the project is no longer actively maintained. James is a pretty cool guy though. Golinks was missing usage statistics and was using a technology stack unknown to me (Bitcask) so I decided to implement my own solution instead of contributing back.
- [Tailscale's golink](https://github.com/tailscale/golink): if you use Tailscale, golink provides short names for your services running in your tailnets. I don't use Tailscale so a few things don't work.
- [Trotto](https://github.com/trotto/go-links): Postgres backed solution, aimed at teams.
- [golinks.io](https://github.com/GoLinks/golinks): An AI-DrIvEn PlAtFoRm to retrieve and share information in seconds with internal short links, called Go Links®. Empowering teams to move faster and remain focused.
- Bunny1: one of the first Golinks manager, not linking because the website has been taken with ads.

## Acknowledgements

- favicon by [Mihimihi](https://www.flaticon.com/free-icon/rabbit_7441511)
