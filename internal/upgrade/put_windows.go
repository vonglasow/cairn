package upgrade

import "os"

func put(from, to string) error {
	aside := to + ".old"
	_ = os.Remove(aside)
	if err := os.Rename(to, aside); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(from, to)
}
