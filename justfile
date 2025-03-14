generate:
    if type tygo > /dev/null; then \
        tygo generate; \
    else \
        docker compose -f docker-compose.generate.yaml up; \
    fi

clean:
    rm -r ts-types/
