# Setup
Set the absolute path that the project resides in.
E.g.
``` bash
export PROJECT_DIR=/path/to/project
```

# Postgres

``` bash
docker run --name yumyums-db -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password  -p 5432:5432 -d postgres
docker exec -ti yumyums-db psql -U postgres -c 'create database "receipts"'
```

# Metabase

``` bash
docker run -d -p 3000:3000 --name metabase metabase/metabase
```

Open a browser and go to http://localhost:3000

## Dev
Connect to metabase to postgres instance running on host port 5432 via `host.docker.internal`
