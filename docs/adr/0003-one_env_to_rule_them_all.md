# TITLE

Date: 2024-10-27

## Status

Accepted

## Context

Configuration of applications can be very difficult. 

In particular, having applications configure themselves consistently in different environments (e.g. Docker, CF) can mean for errors between the environments.

## Decision

To keep things simple, everything will be through one, centralized configuration file. As much as possible, individual applications will not override this via the command-line (for simplicity). The config defines the application's runtime environment.

As much as possible, the local environment will be made to mirror that which will be encountered in the CF environment.

## Consequences

It could be annoying to remove command-line options, but they don't make sense in the CF env, so they shouldn't matter locally.

The config could get large, but it can be (if needed) broken up into multiple files. (Or, ultimately, be rewritten in [Jsonnet](https://jsonnet.org/) and rendered to JSON.) 

## References

* Jsonnet in go: https://github.com/google/go-jsonnet