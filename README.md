# hashtag_poison

[![Software
License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/mannkind/hashtag_poison/blob/master/LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/mannkind/hashtag_poison)](https://goreportcard.com/report/github.com/mannkind/hashtag_poison)

A silly attempt to poison Twitter's trending topic nonsense

## Installation

* git clone https://github.com/mannkind/hashtag_poison
* cd hashtag_poison
* make
* ./bin/hashtag_poison

## Configuration

Configuration happens in the `config.yaml` file. A full example might look this:

```
---
Accounts:
 - ConsumerKey: "D"
   ConsumerSecret: "C"
   OAuthToken: "B"
   OAuthSecret: "A"
   Name: "@example"
```

