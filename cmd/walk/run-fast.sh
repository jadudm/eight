#!/bin/bash

export SERVICE=walk

pushd /home/vcap/app/cmd/${SERVICE}
    echo Running the $SERVICE
    make run
popd