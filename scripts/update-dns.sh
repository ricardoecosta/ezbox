#!/usr/bin/env bash
curl -X GET -s "http://api.dynu.com/nic/update?hostname=ezbox.dynu.net&password=$DYNU_PASSWORD" > /dev/null