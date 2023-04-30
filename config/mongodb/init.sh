#!/bin/bash

# Wait for MongoDB to start up
until mongosh --quiet --host localhost --eval "print(\"waited for connection\")"; do
  echo "waiting until mongodb is up, sleeping for 1 sec..."
  sleep 1
done


# Check if database exists
if ! mongosh --quiet --host localhost --eval 'show dbs' | grep -q 'featurelabdb'; then
  echo "featurelabdb doesn't exist, creating database and collection..."

  echo "Seeding MongoDB..."
  mongoimport --host localhost --db featurelabdb --collection featurelab --type json --file /app/featurelab/db/config/data.json --jsonArray

  echo "Creating index..."
  mongosh --quiet --host localhost --eval 'db.featurelab.createIndex({ app: 1, feature: 1 }, { unique: true });'

  echo "Finished creating featurelabdb"
else
  echo "Database and collection already exist"
fi