# GoMarks

![](https://raw.githubusercontent.com/sebw/gomarks/refs/heads/main/static/favicon.png)

Yet another self-hosted Golinks manager.

Create smart shortcuts/bookmarks/aliases and make GoMarks your default search engine to speed up your browsing experience.

Similar to Duckduckgo <code>!bangs</code> or the multitude of self-hosted Golinks manager (see below for some alternatives).

Bangs are pretty cool but it means requests always transit through Duckduckgo.

## Features

- can run with Docker, Podman or Kubernetes or as a standalone binary
- can run locally or anywhere (read security section!)
- can act as your default search engine in Firefox, Chrome or even iPhone. No DNS or local host files tweaking needed.
- simple shortcuts to redirect to websites (example: bbc takes you to BBC website)
- smart shortcuts using placeholder <code>%s</code> to redirect to websites with search engines (example: "amazon rasberry pi 5" takes you immediately to Amazon's results for Raspberry)
- single word placeholder. `docker alpine` takes you to Docker Hub but `docker version openshift` takes you to your preferred search engine.
- if your query doesn't match any smart link, your query is sent to your preferred search engine
- shortcuts usage statistics
- reset statistics per shortcut or all
- queries history
- possibility to delete history
- configuration is stored in a sqlite database, allowing to manipulate and backup your data easily

## But Why?

I wanted to learn the basics of Go over a week-end and the Golinks manager I've been using for 3 years was lacking a few features I wanted.

Oh, you wanted to know why the logo is a bunny? Historically one of the first Golinks manager was called bunny1. But mostly, I have 3 pet rabbits!

## Docker Installation 

```bash
docker run -d --name gomarks --restart unless-stopped -v /opt/docker/gomarks:/data -p 8080:8080 ghcr.io/sebw/gomarks:latest
```

Your GoMarks instance runs at `http://localhost:8080`.

Go to the help section for instructions.

## Build Your Own Image

```bash
docker build -f Dockerfile -t gomarks:0.1
```

## Security

There's no authentication, users management or certificates. 

An internet exposed gomarks can be used maliciously.

You MUST secure GoMarks behind things like Let's Encrypt, Authentik, Authelia or Cloudflare.

You can follow [this guide](https://blog.wains.be/2023/2023-01-07-cloudflare-zero-trust-authentik/) to secure GoMarks (SSO + HTTPS) behind Cloudflare and Authentik.

## Screenshots

![](https://raw.githubusercontent.com/sebw/gomarks/refs/heads/main/screenshots/index.png)

![](https://raw.githubusercontent.com/sebw/gomarks/refs/heads/main/screenshots/edit.png)

![](https://raw.githubusercontent.com/sebw/gomarks/refs/heads/main/screenshots/fallback.png)

![](https://raw.githubusercontent.com/sebw/gomarks/refs/heads/main/static/help/chrome_step1.png)

![](https://raw.githubusercontent.com/sebw/gomarks/refs/heads/main/static/help/firefox_step1.png)

![](https://raw.githubusercontent.com/sebw/gomarks/refs/heads/main/static/help/iphone_step1.jpg)

![](https://raw.githubusercontent.com/sebw/gomarks/refs/heads/main/static/help/iphone_step2.jpg)

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
- Bunny1: one of the first Golinks manager, not linking because the website has been taken over by ads. Seriously, don't go to that website.

## Acknowledgements

- favicon by [Mihimihi](https://www.flaticon.com/free-icon/rabbit_7441511)
