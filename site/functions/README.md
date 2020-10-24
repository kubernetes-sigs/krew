# Dynamic functions on Krew documentation

Krew site makes use of Netlify Functions (lambdas) to fetch plugin list
dynamically from `krew-index` repository using GitHub API.

## Set up on Netlify

Functions require a one-time set up in the Netlify console.

Regarding **GitHub API rate limits**:

- In production, make **sure to set  a `GITHUB_ACCESS_TOKEN` environment
  variable** with a permissionless "personal access token" to elevate the rate
  limits for our functions.

  Dynamic responses from the functions are cached on
  Netlify’s CDN for a long time, so this is not a huge problem and a single
  token is very likely to suffice a long period of time.

- During local development, you can hit the GitHub API rate limit as well.
  You should set the  `GITHUB_ACCESS_TOKEN` environment variable as needed.

## Local development

Start `hugo` local iteration server on port 1313:

```
cd ./site
hugo serve
```

In another terminal window, build and start the functions server on port 8080:

```
cd ./functions
go run ./server -port=8080
```

Now, you can reach the website at http://localhost:8080 and the functions at
paths defined in the code e.g.
http://localhost:8080/.netlify/functions/api/plugins.
