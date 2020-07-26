package pipeline

import (
	"image"
	"sync"

	"github.com/nfnt/resize"
)

const (
	sizeSmall  = 64
	sizeMedium = 128
	sizeLarge  = 256
)

// Pipeline is an access point to the pipeline
type Pipeline struct {
	jobs chan *Job
	wg   sync.WaitGroup
	done chan struct{}
}

// Job keeps necessary info for executing a job in pipeline
type Job struct {
	UUID        string
	Read        func() (image.Image, error)
	WriteSmall  func(image.Image) error
	WriteMedium func(image.Image) error
	WriteLarge  func(image.Image) error
	Callback    func(*Job)
	Err         error

	image image.Image
}

type thumbJob struct {
	*Job
	write     func(image.Image) error
	thumbnail image.Image
}

// Boot starts a a pipeline with input buffer of inputSize capacity
// and pipelinesCount concurrent pipelines.
// Returns an pointer to the Pipeline representation.
func Boot(inputSize, pipelinesCount int) *Pipeline {
	p := &Pipeline{}
	for i := 0; i < pipelinesCount; i++ {
		buildPipeline(p.jobs, p.done, &p.wg)
	}
	return p
}

// Push a Job into the pipeline queue
func (p *Pipeline) Push(job *Job) {
	p.jobs <- job
}

// Shutdown the Pipeline. Blocks until all pipeline goroutines exits.
func (p *Pipeline) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
}

func buildPipeline(input <-chan *Job, done <-chan struct{}, wg *sync.WaitGroup) {
	origPicture := make(chan *Job)
	smallWorker := make(chan *thumbJob)
	mediumWorker := make(chan *thumbJob)
	largeWorker := make(chan *thumbJob)
	writerWorker := make(chan *thumbJob)
	output := make(chan *Job)

	wg.Add(5)
	go readImage(wg, input, origPicture)
	go fanOut(wg, origPicture, smallWorker, mediumWorker, largeWorker)

	thumbnizeWG := &sync.WaitGroup{}
	thumbnizeWG.Add(3)
	go thumbnize(sizeSmall, thumbnizeWG, smallWorker, writerWorker)
	go thumbnize(sizeMedium, thumbnizeWG, mediumWorker, writerWorker)
	go thumbnize(sizeLarge, thumbnizeWG, largeWorker, writerWorker)
	go func() {
		thumbnizeWG.Wait()
		close(writerWorker)
		wg.Done()
	}()

	go writer(wg, writerWorker, output)
	go gatherResults(wg, output)
}

func readImage(wg *sync.WaitGroup, in <-chan *Job, out chan<- *Job) {
	defer wg.Done()
	defer close(out)
	var err error
	for job := range in {
		job.image, err = job.Read()
		if err != nil {
			job.Err = err
		}
		out <- job
	}
}

func fanOut(wg *sync.WaitGroup, in <-chan *Job, outs, outm, outl chan<- *thumbJob) {
	defer wg.Done()
	defer close(outl)
	defer close(outm)
	defer close(outs)
	for job := range in {
		if job.Err != nil {
			job.Callback(job)
			continue
		}
		outs <- &thumbJob{job, job.WriteSmall, nil}
		outm <- &thumbJob{job, job.WriteMedium, nil}
		outl <- &thumbJob{job, job.WriteLarge, nil}
	}
}

func thumbnize(size uint, wg *sync.WaitGroup, in <-chan *thumbJob, out chan<- *thumbJob) {
	defer wg.Done()
	for job := range in {
		if job.Err == nil {
			job.thumbnail = resize.Thumbnail(size, size, job.image, resize.Lanczos3)
		}
		out <- job
	}
}

func writer(wg *sync.WaitGroup, in <-chan *thumbJob, out chan<- *Job) {
	defer wg.Done()
	defer close(out)
	for job := range in {
		if job.Err == nil {
			err := job.write(job.image)
			if err != nil {
				job.Err = err
			}
		}
		out <- job.Job
	}
}

func gatherResults(wg *sync.WaitGroup, in <-chan *Job) {
	defer wg.Done()
	var wgCache map[string]*sync.WaitGroup
	for job := range in {
		jobWG, ok := wgCache[job.UUID]
		if !ok {
			jobWG = &sync.WaitGroup{}
			jobWG.Add(3)
			wgCache[job.UUID] = jobWG
			go func(job *Job) {
				jobWG.Wait()
				delete(wgCache, job.UUID)
				job.Callback(job)
			}(job)
		}
		jobWG.Done()
	}
}
