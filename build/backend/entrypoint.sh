#!/usr/bin/env bash

goose -dir migrations postgres "$POSTGRES_CONNECTION_URL" up

/app/backend