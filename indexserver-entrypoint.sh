#!/bin/sh
set -e

envsubst < .netrc.template > .netrc
envsubst < config.json.template > config.json
envsubst < gitlab_token.txt.template > gitlab_token.txt

exec "$@"