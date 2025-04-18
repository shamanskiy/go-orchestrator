package schedulers

type Scheduler interface {
	SelectCandidateNodes()
	Score()
	Pick()
}
