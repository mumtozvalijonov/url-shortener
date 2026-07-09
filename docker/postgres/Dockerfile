FROM postgres:18.4-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        postgresql-18-cron \
        postgresql-18-partman \
    && rm -rf /var/lib/apt/lists/*
