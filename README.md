# eight

Experiment [six](https://github.com/jadudm/six) explored the creation of an end-to-end web crawler/indexer/search server in 2000 lines of Go.

Experiment eight is about getting some details closer to "right."

## running the experiment

In the top directory, first build the base/build container:

```
make build
```

Then, run the stack.

```
make run
```

The latter is just `docker compose up`

## interacting with components

To fetch a page from the net, and store it to S3:

```
http PUT localhost:10000/fetch host=cloud.gov path=/pages
```

### browsing the local S3 store

[Minio](https://min.io) is used to simulate S3 locally. 

![alt text](docs/images/minio.png)

Point a browser at [localhost:9001](http://localhost:9001) with the credentials `nutnutnut/nutnutnut` to browse.

### watching the queue

There is a UI for monitoring the queues.

![alt text](docs/images/riverui.png)

This lets you watch the queues at [localhost:11111](http://localhost:11111) provided by [River](https://riverqueue.com/), a queueing library/system built on Postgres. 

## star history

I saw this on another project and thought it was cute. Here, it might be ironic.

[![Star History Chart](https://api.star-history.com/svg?repos=jadudm/eight&type=Date)](https://star-history.com/#jadudm/eight&Date)
