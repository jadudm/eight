# Use what's avaialble

Date: 2024-10-27

## Status

Accepted

## Context

When developing code, it is possible to do *almost anything*.


![https://www.xkcd.com/353/](https://imgs.xkcd.com/comics/python.png)

However, some things take more time than is likely to exist. 

## Decision

We will, in developing search, initially make pragmatic choices about what we can and cannot do. This will generally manifest as "use what's avaialble."

For example, if there is a high-quality text analytics package for the English language that does not have an analogue in other languages, we will use what exists. (Words like "colonialism" or "inequity" might apply here.) We will make the pragmatic choice (initially) and use the best tools at hand, and later, investigate what is involved in developing fundamental/foundational resources at a later point.

## Consequences

Not all languages are equally supported in NLP tooling. Our architectural choices allow us to use nearly any programming language or tooling in our stack; as much as possible, we hope that good architectural choices will not further exacerbate relative linguistic support in the NLP-world. 