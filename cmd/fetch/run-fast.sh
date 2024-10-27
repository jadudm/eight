#!/bin/bash

export SERVICE=fetch

pushd /home/vcap/app/cmd/${SERVICE}
    echo Running the $SERVICE
    make run
popd