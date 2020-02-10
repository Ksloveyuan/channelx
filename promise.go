package channelx

// Represent the promise
type Promise struct {
	done chan struct{}
	res  interface{}
	err  error
}

// Represent the work function
type WorkFunc func() (interface{}, error)

// Represent the success handler
type SuccessHandler func(interface{}) (interface{}, error)

// Represent the error handler
type ErrorHandler func(error) interface{}

// Create a new promise
func NewPromise(workFunc WorkFunc) *Promise {
	promise := Promise{done: make(chan struct{})}
	go func() {
		defer close(promise.done)
		promise.res, promise.err = workFunc()
	}()
	return &promise
}

// Wait until the promise is done and get response.
func (p *Promise) Done() (interface{}, error) {
	<-p.done
	return p.res, p.err
}

// Chain the promise with success handler and error handler
func (p *Promise) Then(successHandler SuccessHandler, errorHandler ErrorHandler) *Promise {
	newPromise := &Promise{done: make(chan struct{})}
	go func() {
		res, err := p.Done()
		defer close(newPromise.done)
		if err != nil {
			if errorHandler != nil {
				newPromise.res = errorHandler(err)
			} else {
				newPromise.err = err
			}
		} else {
			if successHandler != nil {
				newPromise.res, newPromise.err = successHandler(res)
			} else {
				newPromise.res = res
			}
		}
	}()

	return newPromise
}

// Chain the promise with success handler
func (p *Promise) ThenSuccess(successHandler SuccessHandler) *Promise {
	return p.Then(successHandler, nil)
}

// Chain the promise with error handler
func (p *Promise) ThenError(errorHandler ErrorHandler) *Promise {
	return p.Then(nil, errorHandler)
}
