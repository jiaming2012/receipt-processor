
# Postgres

``` bash
docker run --name yumyums-db -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password  -p 5432:5432 -d postgres
docker exec -ti yumyums-db psql -U postgres -c 'create database "receipts"'
```