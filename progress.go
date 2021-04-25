package wildcat

import (
	"context"
	"sync"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
	"golang.org/x/sync/semaphore"
)

type Progress interface {
	UpdateTarget()
	Wait()
	Done()
}

func initProgressBar(progress Progress) *ProgressBar {
	p := mpb.New(mpb.WithWidth(64))
	bar := p.Add(0,
		mpb.NewBarFiller(""),
		mpb.PrependDecorators(
			decor.Counters(0, "% 2d/% 2d "),
			// display our name with one space on the right
			decor.Name("targets"),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
	return &ProgressBar{progress: progress, total: 0, bar: bar, mpb: p}
}

func NewProgress(showBar bool, max int64) Progress {
	group := new(sync.WaitGroup)
	var progress Progress = &nullProgress{group: group}
	if max > 0 {
		progress = &limittedProgress{group: group, sem: semaphore.NewWeighted(max)}
	}
	if showBar {
		progress = initProgressBar(progress)
	}
	return progress
}

type ProgressBar struct {
	mpb      *mpb.Progress
	bar      *mpb.Bar
	total    int64
	progress Progress
}

func (pb *ProgressBar) Wait() {
	pb.progress.Wait()
	pb.bar.SetTotal(pb.total, true)
	pb.mpb.Wait()
}

func (pb *ProgressBar) UpdateTarget() {
	pb.progress.UpdateTarget()
	pb.total = pb.total + 1
	pb.bar.SetTotal(pb.total, false)
}

func (pb *ProgressBar) Done() {
	pb.progress.Done()
	pb.bar.Increment()
}

type limittedProgress struct {
	sem   *semaphore.Weighted
	group *sync.WaitGroup
}

func (lp *limittedProgress) Wait() {
	lp.group.Wait()
}

func (lp *limittedProgress) UpdateTarget() {
	lp.sem.Acquire(context.Background(), 1)
	lp.group.Add(1)
}

func (lp *limittedProgress) Done() {
	lp.sem.Release(1)
	lp.group.Done()
}

type nullProgress struct {
	group *sync.WaitGroup
}

func (np *nullProgress) Wait() {
	np.group.Wait()
}

func (np *nullProgress) UpdateTarget() {
	np.group.Add(1)
}

func (np *nullProgress) Done() {
	np.group.Done()
}
