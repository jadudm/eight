# Use the queues

Date: 2024-10-26

## Status

Accepted

## Context

The goal is to have a scalable search system. This implies a number of services, many of which are unknown at the start. We don't even know if all of the services will be built in the same language.

Queues are a common/natural data structure for an environment where the creation of jobs may come from many (asynchronous) sources, and the worker(s) are available to do the work as discrete, decoupled systems.

## Decision

To decentralize and decouple the services, we will use queues as much as possible for communication between services. 

## Consequences

There are concerns and consequences with queueing systems, especially regarding job completion. An important design consequence is that individual systems/components must operate such that the job they are doing can be done again without consequence in the event that they crash before reporting that their work is complete.

## References

In Sampson's pattern language, this would be a __Farm__ pattern, and we consider a queue to be equivalent to a shared, buffered channel in CSP. 


* Sampson, [Process-Oriented Patterns For Concurrent Software Engineering](https://offog.org/publications/ats-thesis.pdf#page=131)
* Fowler. [Catalog of Patterns of Distributed Systems
](https://martinfowler.com/articles/patterns-of-distributed-systems/)
* O'Reilly. [Designing Distributed Systems](https://info.microsoft.com/rs/157-GQE-382/images/EN-CNTNT-eBook-DesigningDistributedSystems.pdf#page=123)