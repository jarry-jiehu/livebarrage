package comet

type Worker struct {
	id         int
	ctrl       chan bool
	jobChannel chan Job
}

type Job interface {
	Do(int) bool
}

func NewWorker(id int, channelSize int) *Worker {
	return &Worker{
		id:         id,
		ctrl:       make(chan bool),
		jobChannel: make(chan Job, channelSize),
	}
}

func (this *Worker) GetChannel() *chan Job {
	return &this.jobChannel
}

func (this *Worker) Run() {
	go func() {
		for {
			select {
			case job := <-this.jobChannel:
				job.Do(this.id)
			case <-this.ctrl:
				return
			}
		}
	}()
}

func (this *Worker) Stop() {
	close(this.ctrl)
}
