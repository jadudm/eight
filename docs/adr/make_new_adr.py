import click
import os
import shutil


@click.command()
@click.argument("name")
def main(name):
    files = os.listdir(os.curdir)
    adrs = 0
    for f in files:
        if f.endswith("md"):
            adrs += 1
    adrs += 1
    shutil.copyfile("TEMPLATE", f"{adrs:04}-{name}.md")


if __name__ in "__main__":
    main()
