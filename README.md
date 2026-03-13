# BootCrawler

Web crawler that generates a JSON report from a given website.

## Requirements

- go (1.26.1+)

## Installation

```bash
git clone https://github.com/el-damiano/bootdev-crawler.git &&
cd bootdev-crawler &&
go build .
```

## Usage

```bash
./bootdev-crawler [URL] [max concurrency] [max pages]
```

`[URL]`  specifies the website where crawling starts.

`[max concurrency]`  specifies the number of jobs to run concurrently.

`[max pages]`  specifies the limit of pages after which to stop, useful for
crawling websites with thousands of pages.
