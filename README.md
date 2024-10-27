# eight

Experiment [six](https://github.com/jadudm/six) explored the creation of an end-to-end web crawler/indexer/search server in 2000 lines of Go.

Experiment eight is about getting some details closer to "right."

## goals / design

There are a bunch of possible goals. Some might be incompatible.

### compliance-first

The goal is a service that can achieve ATO. That means a whole host of things about deployment, access control, logging, etc.

#### FISMA?

Is this a search engine that operates at FISMA Low, Medium, or High? This question will impact some issues of overall service design. For example, we may want to keep separate things separate if we're operating at High. This could be an argument for a "search engine in a box" (next). Systems that want to be clean/separate can just be their own instances.

### single-server, end-to-end

Do I want a "search engine in a box?" That is, do I want to be able to deploy a single server instance, with nothing more than a connection to S3, and be able to run a complete search service? Maybe. That would be an interesting design constraint.

Working with this, though, it suggests that library/module design should be carried out in a way that _either_ many small services can be built, _or_ a single service that combines them all. 

### live vs. static

There is one world where the search component is live; that is, a page might be querying an app on our infra, and receiving results.

There is another world where the result of the indexing and content cleanup is a static asset that can be embedded in a static site build. 

There might be a third world where all of this runs on someone else's infrastructure to produce the assets in question---dynamic or otherwise. 

For now, being able to handle an end-to-end process that yields a living search engine, and knowing that it can also generate a static site search as well.

Possible static site tools:

* [tinysearch](https://github.com/tinysearch/tinysearch)
* [pagefind](https://pagefind.app/)

and the Hugo project maintains a [list of more](https://gohugo.io/tools/search/).

### search is a data pipeline

First you need to crawl a site, and grab the content.

Then you need to process that content. Perhaps you index it. Perhaps you apply AI to images to determine if there are cats present. Or cats eating hotdogs. Or dogs and cats, living together.

Then you need to bundle it up into a search interface...

Then you need to track and store usage and performance...

As much as possible, each service/step will consume some content and produce some content. Ideally, all of this scales embarrasingly: meaning, we have jobs on queues, and can throw more workers at the queues if we need things to go faster. The content consumed, once fetched from the web, is shuffled in and out of S3 buckets.

### extensible

Everyone wants that. But, if we hold hard-and-fast to a worker/queue model, treat everything as a pipeline, and develop services in a manner that they are pluggable, it becomes possible to imagine having a base service, and then have more advanced services that come at a cost (because, perhaps, they require more resource to devleop, maintain, serve, etc.). 

AWS did this by saying "everything is an API." Much the same here; common APIs and queueing models (ideally with models that can be accessed from multiple languages, so components can be built in whatever tooling makes the most sense) will be the path to extensibility.

## common component architecture

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

### push vs. pull?

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

## crawler

The crawler needs to:

1. Dequeue pages from a frontier queue
2. Store pages in S3
3. Enqueue work for an indexer or other services

The storage medium could be abstracted away. That is, we might want to store to Postgres _or_ S3. But, there comes a point where you run out of turtles. So, in the spirit of building a notional data pipeline of services that want to do things to content, S3 fees _just fine_.

