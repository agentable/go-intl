package conformance

import (
	"time"
)

func SkipReason(root, id string, now time.Time) (string, bool) {
	divergenceIDs, err := loadDivergenceIDs(divergenceLedgerPath(root))
	if err == nil {
		if _, ok := divergenceIDs[id]; ok {
			return "divergence listed in " + divergenceLedgerRelativePath(), true
		}
	}
	xfails, err := loadXFails(xfailPath(root))
	if err != nil {
		return "", false
	}
	for _, xfail := range xfails {
		if xfail.ID != id {
			continue
		}
		expired, err := xfail.expired(now)
		if err != nil || expired {
			return "", false
		}
		return xfail.Reason, true
	}
	return "", false
}
