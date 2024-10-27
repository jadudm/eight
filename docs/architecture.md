
# common component architecture

Each service should have a common architecture. This way, it is easy to either:

1. Wrap the service as a standalone application, fed by a queue, or 
2. Bundle multiple services in a single application, perhaps with a self-contained queue feeding the components.

In Go, the natural choice is a channel interface.

```mermaid
flowchart LR
  Queue["Queue watcher"] --> Component
```

The queue watcher is a common component that pulls jobs from a queue, and messages them to the component. The channel interface allows for any number of patterns; the component can have a single processing loop that blocks on each job, or it can have a pool of workers processing the jobs that come in. What the component does is hidden from view, and allows for appropriate levels of complexity as needed.

In other contexts, this would be called an "API."

## push vs. pull?

I worry about queue polling. But, I think I worry about pub/sub and its interaction wtih worker pools more. 

That said, I won't rule out `mosquitto`. Or [mochi](https://github.com/mochi-mqtt/server).

(I don't yet have estimates in mind for how much work this system will be doing. Polling might not be an issue.)

# services and components

There are a clear, core set of services and components required.

## storage

Storage will be via S3. For development, a containerized version of Minio. "Same difference," as they say.

The assumption is that services will have no more than 5GB of local storage available to them. They can use it, but it is ephemeral (it goes away with each restart), and it is capped (we only have 5GB).

## queue server

The queue server can be abstracted over any number of implementations. That is, the substrate can be in-memory, SQLite files, Postgres, or Redis. The interface---the kinds of operations that are supported---is what matters.

Experiment eight is using [River](https://riverqueue.com/) as its job/queue. It has a complete Go library, and is built on top of Postgres.

The authors' mindset is "we know how to work with Postgres." From a service design perspective, my goal is to minimize the number of components involved. For now, what if I can get away with *only* having Postgres.

(I would like to ask the question "can I get away with only having SQLite?" I will come back to this question. That way, I'd have no external service dependencies to manage, and backups are just a push to S3.)

## fetch

`fetch` needs to:

1. Dequeue pages from a frontier queue
2. Store pages in S3
3. Enqueue work for an indexer or other services

The storage medium could be abstracted away. That is, we might want to store to Postgres _or_ S3. But, there comes a point where you run out of turtles. So, in the spirit of building a notional data pipeline of services that want to do things to content, S3 fees _just fine_.
