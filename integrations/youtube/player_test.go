package youtube

import (
	"fmt"
	"testing"
)

func TestQueue_Pop(t *testing.T) {
	queue := Queue{}

	queue.add(QueueElement{
		QueueElement: nil,
		Packets:      nil,
	})
	fmt.Println(len(queue))
	queue.Pop()
	fmt.Println(len(queue))
}
