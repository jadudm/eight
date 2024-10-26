# crawler

```mermaid
sequenceDiagram
  autonumber
  participant Q as Queue
  participant F as Fetch
  participant C as Cache
  # ---
  Q->>F: get job
  note right of Q: host, path 
  F->>C: check cache
  activate C
  C->>F: response
  deactivate C
  note right of F: count bytes
  note right of F: hash
  note right of F: store to S3
  F->>C: update cache
  F-->>Q: virus scan
  F-->>Q: index[content type]
```

## checking the cache

The cache holds pairs of

```
[host/path] <-> [S3 path]
```

If we check, and `""` comes back, it isn't in the cache.


## libraries

* https://github.com/tidwall/sjson
* https://github.com/tidwall/gjson