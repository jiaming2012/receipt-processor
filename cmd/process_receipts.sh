#!/bin/bash

# Error if the env PROJECT_DIR is not set
if [ -z "$PROJECT_DIR" ]
then
  echo "Error: PROJECT_DIR is not set"
  exit 1
fi

export DATABASE_URL="postgresql://myuser:test123@us.loclx.io:29453/mydb"
echo "hey"
go run $PROJECT_DIR/cmd/process_receipts/main.go