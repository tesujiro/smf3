#!/bin/bash

scan()
{
        local ID=$1

        tile38-cli<<EOF
SCAN flyer
EOF
}

while :
do
        scan | jq
        sleep 2
done
