package utils

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
)

type User struct {
	Gid int
	Uid int
}

func (from User) UpdateGidUid(ctx Context, to User) error {
	if to.Uid == 0 {
		return fmt.Errorf("refusing to update spt user to uid 0")
	}

	if from.Uid != to.Uid {
		ctx.Logger().Info("change uid", "from", from.Uid, "to", to.Uid)
		_, err := RunCommand(ctx, []string{"usermod", "-u", strconv.Itoa(from.Uid), strconv.Itoa(to.Uid)}, CmdOpts{})
		if err != nil {
			return err
		}
	}

	if from.Gid != to.Gid {
		ctx.Logger().Info("change gid", "from", from.Gid, "to", to.Gid)
		_, err := RunCommand(ctx, []string{"groupmod", "-g", strconv.Itoa(from.Gid), strconv.Itoa(to.Gid)}, CmdOpts{})
		if err != nil {
			return err
		}
	}

	return nil
}

func UserFromCurrent(ctx Context) (User, error) {
	fail := func(err error) (User, error) {
		return User{}, err
	}

	user, err := user.Current()
	if err != nil {
		return fail(err)
	}

	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return fail(err)
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return fail(err)
	}

	return User{Gid: gid, Uid: uid}, nil
}

func UserFromEnv(ctx Context, envUid string, envGid string) (User, error) {
	fail := func(err error) (User, error) {
		return User{}, err
	}

	gidString := os.Getenv(envGid)
	if gidString == "" {
		return fail(fmt.Errorf("env var %s unset", envGid))
	}
	gid, err := strconv.Atoi(gidString)
	if err != nil {
		return fail(err)
	}

	uidString := os.Getenv(envUid)
	if uidString == "" {
		return fail(fmt.Errorf("env var %s unset", envGid))
	}
	uid, err := strconv.Atoi(uidString)
	if err != nil {
		return fail(err)
	}

	return User{Gid: gid, Uid: uid}, nil
}

func UserFromUsername(ctx Context, username string) (User, error) {
	fail := func(err error) (User, error) {
		return User{}, err
	}

	lookupUser, err := user.Lookup(username)
	if err != nil {
		return fail(err)
	}

	gid, err := strconv.Atoi(lookupUser.Gid)
	if err != nil {
		return fail(err)
	}

	uid, err := strconv.Atoi(lookupUser.Uid)
	if err != nil {
		return fail(err)
	}

	return User{Gid: gid, Uid: uid}, nil
}
