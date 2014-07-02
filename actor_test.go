package actor

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
)

type MyMsg struct {
	GenericMsg
	val int
}

func (msg *MyMsg) Type() MsgType {
	return "actor.MyMsg"
}

type MyHandler struct{}

const (
	kHandlers int = 10
	kMsgs         = 10000
)

func (h *MyHandler) Map(m Msg, c Context) MapSet {
	return MapSet{{"D", Key(strconv.Itoa(m.(*MyMsg).val % kHandlers))}}
}

var testStageCh chan interface{} = make(chan interface{})

func (h *MyHandler) Recv(m Msg, c RecvContext) {
	hash := m.(*MyMsg).val % kHandlers
	d := c.State().Dict("D")
	k := Key(strconv.Itoa(hash))
	v, ok := d.Get(k)
	if !ok {
		v = 0
	}
	i := v.(int) + 1
	d.Set(k, i)

	id := c.(*recvContext).rcvr.id().RcvrId % uint32(kHandlers)
	if id != uint32(hash) {
		panic(fmt.Sprintf("Invalid message %v received in %v.", m, id))
	}

	if i == kMsgs/kHandlers {
		testStageCh <- true
	}
}

func TestStage(t *testing.T) {
	runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(1)

	testStageCh = make(chan interface{})
	defer func() {
		close(testStageCh)
		testStageCh = nil
	}()

	stage := NewStage()
	actor := stage.NewActor("MyActor")
	actor.Handle(MyMsg{}, &MyHandler{})

	stageWCh := make(chan interface{})
	go stage.Start(stageWCh)

	for i := 1; i <= kMsgs; i++ {
		stage.Emit(&MyMsg{val: i})
	}

	for i := 0; i < kHandlers; i++ {
		<-testStageCh
	}

	stage.Stop()
	<-stageWCh
}
