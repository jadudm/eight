# eight

Experiment [six](https://github.com/jadudm/six) explored the creation of an end-to-end web crawler/indexer/search server in 2000 lines of Go.

Experiment eight is about getting some details closer to "right."

## running the experiment

In the top directory, first build the base/build container:

```
make docker
```

Then, run the stack.

```
make up
```

The `up` compiles all of the services, generates the database API, unpacks the USWDS assets into place, and launches the stack.

## interacting with components

An API key is hard-coded into the `compose.yaml` file; it is `lego`. Obviously, this is not suitable for production use.

To begin a crawl:

```
http PUT localhost:10000/fetch host=fac.gov path=/ api_key=lego
```

The file `container.yaml` is a configuration file with a *few* end-user tunables. The FAC website is small in HTML, but is *large* because it contains 4 PDFs at ~2000 pages each. If you only want to index the HTML, set `extract_pdf` to `false`. (This is good for demonstration purposes.)

To fetch a single PDF and see it extracted:

```
http PUT localhost:10000/fetch host=app.fac.gov path=/dissemination/report/pdf/2023-09-GSAFAC-0000063050
```

(approximately 100 pages)

## searching 

After a site is walked and packed, an SQLite file with full-text capabilities is generated. The `serve` component watches for completed files, grabs them from S3, and serves queries from the resulting SQLite database.

```
http POST localhost:10004/serve  host=fac.gov terms="community grant"
```

is how to search using the API; search terms are a single list, and SQLite pulls them apart.

A [WWW-based search interface](http://localhost:10004/search/) can be found at [http://localhost:10004/search/](http://localhost:10004/search/). You will be redirected to a page that lists all of the sites currently indexed. Note that the final part of the URL

```
http://localhost:10004/search/{HOST}
```

determines what indexed database will be searched. (E.g. if you have indexed `alice.gov` and `bob.gov`, selecting that database will navigate you to a URL like `http://localhost:10004/search/alice.gov`.)

## browsing the backend

The goal is to minimize required services. This stack *only* uses Postgres and S3. 

## browsing S3

The S3 filestore is simulated when running locally using a containerized version of [Minio](https://min.io).

![alt text](docs/images/minio.png)

Point a browser at [localhost:9001](http://localhost:9001) with the credentials `nutnutnut/nutnutnut` to browse.

### watching the queue

There is a UI for monitoring the queues.

![alt text](docs/images/riverui.png)

This lets you watch the queues at [localhost:11111](http://localhost:11111) provided by [River](https://riverqueue.com/), a queueing library/system built on Postgres. 

### observing the database

[pgweb](https://sosedoff.github.io/pgweb/) is included in the container stack for browsing the database directly and, if needed, editing. Pointign a browser at [localhost:22222](http://localhost:22222) will bring up `pgweb`.

If you are running, and want to simulate total queue loss, run

```
truncate table river_job; truncate table river_leader; truncate table river_queue;
```

This will not break the app; it will, however, leave all of the services with nothing to do.

## other utilities

To run the stack without the services (just the backend of `minio`, `postgres`, and the UIs)

```
docker compose -f backend.yaml up
```

To 

## by the numbers

```
 docker run --rm -v ${PWD}:/tmp aldanial/cloc --exclude-dir=assets --fullpath --not-match-d=terraform/zips/* --not-match-d=terraform/app/* --not-match-d=.terraform/* .
```

```
--------------------------------------------------------------------------------
Language                      files          blank        comment           code
--------------------------------------------------------------------------------
Text                              2              0              0          10127
Go                               46            553            250           2628
YAML                              7             25            106            855
Markdown                         13            207              0            347
HTML                              1             43              0            254
JSON                              1              0              0            199
make                              8             41              1            129
Python                            3             18              0             83
Dockerfile                        6             26             15             61
Bourne Shell                      5             10              0             30
SQL                               2              8             10             25
Bourne Again Shell                1              2              3              6
--------------------------------------------------------------------------------
SUM:                             95            933            385          14744
--------------------------------------------------------------------------------
```