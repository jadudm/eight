#!/bin/bash

export SERVICE=serve

pushd /home/vcap/app/cmd/${SERVICE}
    echo Running the $SERVICE
    make run
popd