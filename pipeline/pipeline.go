package pipeline

type Pipeline struct {
	jobs chan Job
	done chan struct{}
}

type Job struct {
	UUID        string
	picturePath string
	outputPath  string
	callback    func(string, string, string)
}

func Boot(inputSize, pipelinesCount int) *Pipeline {
	p := &Pipeline{
		jobs: make(chan Job, inputSize),
		done: make(chan struct{}),
	}
	for i := 0; i < pipelinesCount; i++ {
		buildPipeline(p.jobs, p.done)
	}
	return p
}

func buildPipeline(input <-chan Job, done <-chan struct{}) {
	go func() {
		
	}
}
