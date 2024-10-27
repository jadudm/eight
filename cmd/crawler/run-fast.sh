#!/bin/bash

export SERVICE=crawler

pushd /home/vcap/app/cmd/${SERVICE}
    echo Running the $SERVICE
    make run
popd