default:
    just --list

build: build-ts

generate-ts:
    if type tygo > /dev/null; then \
        tygo generate; \
    else \
        docker compose -f docker-compose.generate.yaml up; \
    fi

build-ts: generate-ts
    npm run build

run: build
    go run .

clean:
    rm -r ts/types/
    npm run clean
