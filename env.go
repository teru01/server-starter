package starter

import "errors"

var errNoEnv = errors.New("no ENVDIR specified, or ENVDIR does not exist")

func reloadEnv() (map[string]string, error) {

}


