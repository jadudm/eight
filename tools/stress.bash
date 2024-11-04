#!/bin/bash

# use by calling
# ./stress.bash N
# where `N` is the number of stressors to run in parallel.
 
for i in $(seq $2)
do
  echo $i
  python stress_the_search.py $1 1000 &
done
