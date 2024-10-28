#!/bin/bash

export SERVICE=pack

pushd /home/vcap/app/cmd/${SERVICE}
    echo Running the $SERVICE
    make run
popd