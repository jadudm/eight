## assets

This is a weird folder.

We don't keep the `assets` around permanently; they're compressed in `static.zip`.

The `assets/index.html` file is what `serve` dishes out. Instead of copying it from that folder on deploy, I'm just editing it in place, as that reduces the amount of work done while testing and deploying.