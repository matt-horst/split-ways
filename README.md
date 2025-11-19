# Split Ways
A web based appliction to track household expenses for a group similar to [splitwise](https://www.splitwise.com).

## Setup
### .env
Create a `.env` file with keys for the desired `PORT`, `DATABASE` connection string, and `JWT_SECRET`. The port will default to `8080`; the database connection string and JWT secret are required.

### Database Migrations
Database migrations are intended to be run automaticall by [goose](https://github.com/pressly/goose). This can be installed by running the following:
```
go install github.com/pressly/goose@latest
```
After installation, the database migrations can be run executing:
```
goose <db_driver> <db_connection_string> -dir ./sql/schema up
```

## Usage
### Start the server
The server can be started with `go run .`.

The front-end should now be accessible from the web-browser at [http://localhost:8080/](http://localhost:8080/), if using the default configuration.

## Development
### SQLC
This package uses generated go code from queries written in sql. Modifying or creating new queries should be generating the corresponing go queries by running `sqlc generate`.

To install `sqlc`, run the following:
```
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## TODO
- [x] Allow editing of expenses 
- [x] Allow editing of payments
- [ ] Add a reset end point for testing
- [ ] Add admin accounts / permissions
- [ ] Add integration tests
- [ ] Implement refresh tokens for auth
- [x] Add link to create group in dashboard
- [ ] Add links to edit / delete groups that the user creates
- [x] Add style to the list of users in add user page
- [ ] Add way to specify how the expense should be split
- [ ] Add dates to the transaction list
- [ ] Setup logging middleware
- [ ] Setup sqlc in CI
- [ ] Add way to remove members from group
- [ ] Add database transaction rollback on multi-step processes
