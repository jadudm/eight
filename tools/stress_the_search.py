import requests as R
import random
import click
from pprint import pprint
import time
from datetime import timedelta

words = []


@click.command()
@click.argument("host")
@click.argument("queries", default=1000)
def maine(host, queries):

    session = R.Session()
    adapter = R.adapters.HTTPAdapter(pool_connections=100, pool_maxsize=100)
    session.mount("http://", adapter)
    start = time.perf_counter()

    for _ in range(0, queries):
        term_count = random.randint(1, 4)
        terms = ""
        for _ in range(term_count):
            ch = random.choice(words)
            ch = ch.strip()
            terms += ch + " "
        session.post(
            "http://localhost:10004/api/search", json={"terms": terms, "host": host}
        )

    duration = timedelta(seconds=time.perf_counter() - start)
    print(f"queries: {queries}")
    print(f"elapsed time: {duration}")
    print(f"time/query: {duration/queries}")

    res = R.get("http://localhost:10004/api/stats")
    pprint(res.json()["stats"])


if __name__ in "__main__":
    with open("random_words.txt") as fp:
        for w in fp:
            if len(w.strip()) > 3:
                words.append(w.strip())
    maine()
