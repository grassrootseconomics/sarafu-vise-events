package lookup

import (
	"git.grassecon.net/grassrootseconomics/sarafu-api/remote/http"
	"git.grassecon.net/grassrootseconomics/sarafu-api/remote"
)

var (
	// Api provides the api implementation for all external lookups.
	Api remote.AccountService = &http.HTTPAccountService{}
)
