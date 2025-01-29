package stacktrace

type BuildStackMode int

const (
	TraceCollect BuildStackMode = iota + 1
	TraceBorrowOrCollect
	TraceBorrowOnly
	TraceEnhance
	TraceOmit
	TraceTrimmed
)
