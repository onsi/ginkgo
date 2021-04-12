// +build windows

package parallel_support

func NewOutputInterceptor() OutputInterceptor {
	return noopOutputInterceptor{}
}
