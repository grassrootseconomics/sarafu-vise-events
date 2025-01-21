package lookup

import (
	"git.grassecon.net/grassrootseconomics/sarafu-api/remote"
	"git.grassecon.net/grassrootseconomics/sarafu-api/remote/http"
)

var (
	// Api provides the api implementation for all external lookups.
	Api remote.AccountService = &http.HTTPAccountService{}
)
