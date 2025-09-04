package caller

import "github.com/dmitastr/yp_gophermart/internal/domain/api_caller/caller/accrualcaller"

type Caller interface {
	AddJob(string) (chan accrualcaller.WorkerResult, error)
}
