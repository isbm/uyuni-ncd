/*
Message ID buffer is used to filter messages from mirroring to itself.
All messages that fire to the same channel are also received back.
So the buffer is used to prevent sender get own prodecural message back.
*/

package ncdtransport

type MsgIdStor struct {
	buff map[string]string
}

func NewMsgIdStor() *MsgIdStor {
	mis := new(MsgIdStor)
	mis.buff = make(map[string]string)
	return mis
}

func (mis *MsgIdStor) Push(msgid string) {
	if !mis.Present(msgid) {
		mis.buff[msgid] = ""
	}
}

// Check if message ID is present in the stor
func (mis *MsgIdStor) Present(msgid string) bool {
	_, ex := mis.buff[msgid]
	return ex
}

func (mis *MsgIdStor) Pop(msgid string) {
	if mis.Present(msgid) {
		delete(mis.buff, msgid)
	}
}

type MsgIdBuff struct {
	buff map[string]MsgIdStor
}

func NewMsgIdBuff() *MsgIdBuff {
	mb := new(MsgIdBuff)
	mb.buff = make(map[string]MsgIdStor)

	return mb
}

func (mb *MsgIdBuff) Channel(channel string) *MsgIdStor {
	if _, ex := mb.buff[channel]; !ex {
		mb.buff[channel] = *NewMsgIdStor()
	}
	stor := mb.buff[channel]
	return &stor
}
