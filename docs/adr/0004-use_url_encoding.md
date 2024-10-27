# Use URL encoding

Date: 2024-10-27

## Status

Accepted

## Context

Sometimes, bytewise-data needs to be encoded as text.

## Decision

Use URL encoding.

## Consequences

None, really. But, we should be consistent. By always choosing URL encoding, the data in question could (if needed) be shipped as part of a GET/PUT/POST/etc.

