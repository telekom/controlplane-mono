package conjur

import (
	"errors"

	"github.com/cyberark/conjur-api-go/conjurapi/response"
)

func AsError(err error) (*response.ConjurError, bool) {
	var cErr *response.ConjurError
	if errors.As(err, &cErr) {
		return cErr, true
	}
	return nil, false
}
