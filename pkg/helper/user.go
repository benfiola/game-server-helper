package helper

import (
	"fmt"
	"os/user"
	"strconv"

	"github.com/caarlos0/env/v11"
)

// User holds gid/uid information about a given user
type User struct {
	Gid int `env:"GID" envDefault:"1000"`
	Uid int `env:"UID" envDefault:"1000"`
}

// Returns a [User] representing the current process' user
// Returns an error if fetching the current user fails.
// Returns an error if the resulting user has a non-numeric gid/uid.
func (api *Api) GetCurrentUser() (User, error) {
	fail := func(err error) (User, error) {
		return User{}, err
	}

	current, err := user.Current()
	if err != nil {
		return fail(err)
	}

	gid, err := strconv.Atoi(current.Gid)
	if err != nil {
		return fail(err)
	}

	uid, err := strconv.Atoi(current.Uid)
	if err != nil {
		return fail(err)
	}

	return User{Gid: gid, Uid: uid}, nil
}

// Looks up a user by username and returns a [User].
// Returns an error if the lookup fails.
// Returns an error if the resulting user has a non-numeric gid/uid.
func (api *Api) LookupUser(username string) (User, error) {
	fail := func(err error) (User, error) {
		return User{}, err
	}

	lookup, err := user.Lookup(username)
	if err != nil {
		return fail(err)
	}

	gid, err := strconv.Atoi(lookup.Gid)
	if err != nil {
		return fail(err)
	}

	uid, err := strconv.Atoi(lookup.Uid)
	if err != nil {
		return fail(err)
	}

	return User{Gid: gid, Uid: uid}, nil
}

// Returns a [User] representing a gid/uid as set in the environment
// Returns an error if the resulting user has a non-numeric gid/uid.
func (api *Api) GetEnvUser() (User, error) {
	user := User{}
	err := env.Parse(&user)
	return user, err
}

// Updates the gid/uid of the given username
func (api *Api) UpdateUser(username string, to User) error {
	if to.Uid == 0 {
		return fmt.Errorf("refusing to update username %s to uid 0", username)
	}

	from, err := api.LookupUser(username)
	if err != nil {
		return err
	}

	if from.Uid != to.Uid {
		api.Logger.Info("change uid", "user", username, "from", from.Uid, "to", to.Uid)
		_, err := api.RunCommand([]string{"usermod", "-u", strconv.Itoa(to.Uid), username}, CmdOpts{})
		if err != nil {
			return err
		}
	}

	if from.Gid != to.Gid {
		api.Logger.Info("change gid", "user", username, "from", from.Gid, "to", to.Gid)
		_, err := api.RunCommand([]string{"groupmod", "-g", strconv.Itoa(to.Gid), username}, CmdOpts{})
		if err != nil {
			return err
		}
	}

	return nil
}
