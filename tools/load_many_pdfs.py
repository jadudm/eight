import requests
import click
import os
import shutil
import random


@click.command()
@click.argument("count")
def main(count):
    count = int(count)
    g = requests.get(
        "https://api.fac.gov/general",
        params={
            "limit": count,
            "offset": random.randint(0, 10000),
            "audit_year": f"eq.2023",
        },
        headers={"x-api-key": os.getenv("FAC_API_KEY"), "accept-profile": "api_v1_1_0"},
    )
    for o in g.json():
        report_id = o["report_id"]
        p = requests.put(
            "http://localhost:10000/fetch",
            json={
                "host": "app.fac.gov",
                "path": f"dissemination/report/pdf/{report_id}",
            },
        )


if __name__ in "__main__":
    main()
