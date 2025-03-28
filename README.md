# Deadlinkr

This tool allows you to scan a website for broken links (dead links), which is crucial for maintaining the quality of a website. Broken links negatively impact user experience and can affect search engine rankings.

## Features

1. Website crawling: Recursively crawl all pages of a domain
2. Link verification: Test each link to see if it returns an error (404, 500, etc.)
3. Filtering by type: Check internal, external, or both links
4. Depth limitation: Control the depth of the crawling
5. Detailed report: Generate a report of the issues found

## Usage
### Commands

```
deadlinkr scan [url] --format=csv/json/html - Scan a complete website
deadlinkr check [url] --format=csv/json/html - Check a single page
```

### Options et flags

```
--depth=N - Limit the depth of crawling
--concurrency=N - Number of simultaneous requests
--timeout=N - Timeout for each request
--ignore-external - Ignore external links
--only-external - Check only external links
--user-agent="string" - Set a custom user-agent
--include-pattern="regex" - Include only URLs matching the pattern
--exclude-pattern="regex" - Exclude URLs matching the pattern
```

## Roadmap

1. Add a "fix" mode to automatically correct broken internal links
2. Integrate an API to suggest alternatives for broken links
3. Add a web server to visualize reports interactively
4. Implement a "watch" mode to continuously monitor a site
5. Add support for authentication (password-protected sites)