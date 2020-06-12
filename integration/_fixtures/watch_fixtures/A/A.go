package A

import "$ROOT_PATH$/B"

func DoIt() string {
	return B.DoIt()
}
