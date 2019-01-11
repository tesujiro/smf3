#!/bin/bash

check_location()
{
        local ID=$1

        tile38-cli<<EOF
GET location $ID
EOF
}

while :
do
        check_location 0
        sleep 0.2
done
