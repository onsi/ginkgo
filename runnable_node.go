package godescribe

type runnableNode struct {
	isAsync      bool
	asyncFunc    func(Done)
	syncFunc     func()
	codeLocation CodeLocation
}
