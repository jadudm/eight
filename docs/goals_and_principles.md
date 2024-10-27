
# goals / principles

There are a bunch of possible goals. Some might be incompatible.

## compliance-first

The goal is a service that can achieve ATO. That means a whole host of things about deployment, access control, logging, etc.

### FISMA?

Is this a search engine that operates at FISMA Low, Medium, or High? This question will impact some issues of overall service design. For example, we may want to keep separate things separate if we're operating at High. This could be an argument for a "search engine in a box" (next). Systems that want to be clean/separate can just be their own instances.

## single-server, end-to-end

Do I want a "search engine in a box?" That is, do I want to be able to deploy a single server instance, with nothing more than a connection to S3, and be able to run a complete search service? Maybe. That would be an interesting design constraint.

Working with this, though, it suggests that library/module design should be carried out in a way that _either_ many small services can be built, _or_ a single service that combines them all. 

## live vs. static

There is one world where the search component is live; that is, a page might be querying an app on our infra, and receiving results.

There is another world where the result of the indexing and content cleanup is a static asset that can be embedded in a static site build. 

There might be a third world where all of this runs on someone else's infrastructure to produce the assets in question---dynamic or otherwise. 

For now, being able to handle an end-to-end process that yields a living search engine, and knowing that it can also generate a static site search as well.

Possible static site tools:

* [tinysearch](https://github.com/tinysearch/tinysearch)
* [pagefind](https://pagefind.app/)

and the Hugo project maintains a [list of more](https://gohugo.io/tools/search/).

## search is a data pipeline

First you need to crawl a site, and grab the content.

Then you need to process that content. Perhaps you index it. Perhaps you apply AI to images to determine if there are cats present. Or cats eating hotdogs. Or dogs and cats, living together.

Then you need to bundle it up into a search interface...

Then you need to track and store usage and performance...

As much as possible, each service/step will consume some content and produce some content. Ideally, all of this scales embarrasingly: meaning, we have jobs on queues, and can throw more workers at the queues if we need things to go faster. The content consumed, once fetched from the web, is shuffled in and out of S3 buckets.

## extensible

Everyone wants that. But, if we hold hard-and-fast to a worker/queue model, treat everything as a pipeline, and develop services in a manner that they are pluggable, it becomes possible to imagine having a base service, and then have more advanced services that come at a cost (because, perhaps, they require more resource to devleop, maintain, serve, etc.). 

AWS did this by saying "everything is an API." Much the same here; common APIs and queueing models (ideally with models that can be accessed from multiple languages, so components can be built in whatever tooling makes the most sense) will be the path to extensibility.