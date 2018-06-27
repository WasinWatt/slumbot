POSTGRES_SERVER=35.198.242.197

localdb:
	dropdb -h localhost -U postgres slumbot --if-exists
	createdb -h localhost -U postgres slumbot
	pg_dump -O -h $(POSTGRES_SERVER) -U postgres slumbot | psql -h localhost -U postgres slumbot