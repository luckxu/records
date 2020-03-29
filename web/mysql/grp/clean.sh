#!/bin/sh -e


function clean() {
    path=`pwd`/node$1
    if [ -d $path ]; then
        cd $path
        docker-compose down
        cd -
        rm -rf $path
    fi
}

for ((i=1;i<=$1;i++)); do
    clean $i
done
