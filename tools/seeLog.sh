

#!/bin/bash

hostAddr=http://195.201.42.209:18003

curl -X POST -F 'username=admin' -F 'password=admin' $hostAddr/login --cookie-jar ./ParsinCookie

file=$(mktemp)
trap 'rm $file' EXIT

(while true; do
    # shellcheck disable=SC2094
    sleep 1
    curl -s --cookie ./ParsinCookie --fail -r "$(stat -c %s "$file")"- "$hostAddr/data/out.html" 2>> "$file "
done) &
pid=$!
trap 'kill $pid; rm $file; rm ./ParsinCookie' EXIT

tail -f "$file"