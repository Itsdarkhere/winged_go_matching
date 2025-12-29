# MakeFile commands for wingedapp
.PHONY: wingedapp-matching-runner wingedapp-di-generate

wingedapp-di-generate:
	go run ./cmd/wingedapp/di/

wingedapp-matching-runner: wingedapp-di-generate
	go run ./cmd/wingedapp/matching_runner/

wingedapp-recreate-db:
	go run ./cmd/wingedapp/recreatedb

wingedapp-migrate: wingedapp-recreate-db
	go run ./cmd/wingedapp/migrate/

wingedapp-recreate-sqlboiler: wingedapp-migrate
	go run ./cmd/wingedapp/sqlboiler/

wingedapp-terraform-testingenv:
	go run ./cmd/wingedapp/terraform_testingenv/

wingedapp-clean-testdbs:
	go run ./cmd/wingedapp/cleantestdbs/

# MakeFile commands for imapp
.PHONY: imapp-recreate-db imapp-migrate imapp-recreate-sqlboiler

imapp-recreate-db:
	go run ./cmd/imapp/recreatedb

imapp-migrate:
	go run ./cmd/imapp/migrate

imapp-recreate-sqlboiler: imapp-recreate-db imapp-migrate
	go run ./cmd/imapp/sqlboiler