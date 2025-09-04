package caller

import "github.com/dmitastr/yp_gophermart/internal/domain/api_caller/caller/accrual_caller"

type Caller interface {
	AddJob(string) (chan accrual_caller.WorkerResult, error)
}
