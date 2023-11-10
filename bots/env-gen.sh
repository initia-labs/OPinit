#!/bin/bash

# convert envrc to env
if [ ! -f .envrc ]; then
    echo ".envrc not exist."
    exit 1
fi

cp .envrc .envrc.tmp

# remove '' in linux
case "$(uname)" in
    Linux*)     sed -i 's/localhost/host.docker.internal/g' .envrc.tmp ;;
    Darwin*)    sed -i '' 's/localhost/host.docker.internal/g' .envrc.tmp ;;
    *)          echo "Unsupported OS"; exit 1 ;;
esac

sed 's/^export //g' .envrc.tmp > .env
rm .envrc.tmp

echo ".env generated."