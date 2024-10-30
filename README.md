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

To fetch a PDF, and see it extracted:

```
http PUT localhost:10000/fetch host=app.fac.gov path=/dissemination/report/pdf/2023-09-GSAFAC-0000063050
```

(approximately 100 pages)

or

```
http PUT localhost:10000/fetch host=fac.gov path=assets/compliance/2024-Compliance-Supplement.pdf
```

(approximately 2100 pages -- the [2CFR200 Appendix XI Compliance Supplement](https://www.fac.gov/assets/compliance/2024-Compliance-Supplement.pdf))

### browsing the local S3 store

[Minio](https://min.io) is used to simulate S3 locally. 

![alt text](docs/images/minio.png)

Point a browser at [localhost:9001](http://localhost:9001) with the credentials `nutnutnut/nutnutnut` to browse.

### watching the queue

There is a UI for monitoring the queues.

![alt text](docs/images/riverui.png)

This lets you watch the queues at [localhost:11111](http://localhost:11111) provided by [River](https://riverqueue.com/), a queueing library/system built on Postgres. 


## by the numbers

```
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              45            490            232           2218
YAML                             4             14             49            363
Markdown                        12            194              0            318
JSON                             1              0              0            199
Text                             1              0              0            127
Dockerfile                       5             22             12             54
make                             7             17              0             52
Python                           2              8              0             44
SQL                              2              8             10             25
Bourne Shell                     4              8              0             24
-------------------------------------------------------------------------------
SUM:                            83            761            303           3424
-------------------------------------------------------------------------------
```