package yproc

import (
	"fmt"
	"net"

	"github.com/AmarsDing/lib/internal/json"

	"github.com/AmarsDing/lib/container/yqueue"
	"github.com/AmarsDing/lib/container/ytype"
	"github.com/AmarsDing/lib/net/ytcp"
	"github.com/AmarsDing/lib/util/yconv"
)

var (
	// tcpListened marks whether the receiving listening service started.
	tcpListened = ytype.NewBool()
)

// Receive blocks and receives message from other process using local TCP listening.
// Note that, it only enables the TCP listening service when this function called.
func Receive(group ...string) *MsgRequest {
	// Use atomic operations to guarantee only one receiver goroutine listening.
	if tcpListened.Cas(false, true) {
		go receiveTcpListening()
	}
	var groupName string
	if len(group) > 0 {
		groupName = group[0]
	} else {
		groupName = gPROC_COMM_DEFAULT_GRUOP_NAME
	}
	queue := commReceiveQueues.GetOrSetFuncLock(groupName, func() interface{} {
		return yqueue.New(gPROC_MSG_QUEUE_MAX_LENGTH)
	}).(*yqueue.Queue)

	// Blocking receiving.
	if v := queue.Pop(); v != nil {
		return v.(*MsgRequest)
	}
	return nil
}

// receiveTcpListening scans local for available port and starts listening.
func receiveTcpListening() {
	var listen *net.TCPListener
	// Scan the available port for listening.
	for i := gPROC_DEFAULT_TCP_PORT; ; i++ {
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", i))
		if err != nil {
			continue
		}
		listen, err = net.ListenTCP("tcp", addr)
		if err != nil {
			continue
		}
		// Save the port to the pid file.
		if err := yfile.PutContents(getCommFilePath(Pid()), yconv.String(i)); err != nil {
			panic(err)
		}
		break
	}
	// Start listening.
	for {
		if conn, err := listen.Accept(); err != nil {
			ylog.Error(err)
		} else if conn != nil {
			go receiveTcpHandler(ytcp.NewConnByNetConn(conn))
		}
	}
}

// receiveTcpHandler is the connection handler for receiving data.
func receiveTcpHandler(conn *ytcp.Conn) {
	var result []byte
	var response MsgResponse
	for {
		response.Code = 0
		response.Message = ""
		response.Data = nil
		buffer, err := conn.RecvPkg()
		if len(buffer) > 0 {
			// Package decoding.
			msg := new(MsgRequest)
			if err := json.UnmarshalUseNumber(buffer, msg); err != nil {
				//ylog.Error(err)
				continue
			}
			if msg.RecvPid != Pid() {
				// Not mine package.
				response.Message = fmt.Sprintf("receiver pid not match, target: %d, current: %d", msg.RecvPid, Pid())
			} else if v := commReceiveQueues.Get(msg.Group); v == nil {
				// Group check.
				response.Message = fmt.Sprintf("group [%s] does not exist", msg.Group)
			} else {
				// Push to buffer queue.
				response.Code = 1
				v.(*yqueue.Queue).Push(msg)
			}
		} else {
			// Empty package.
			response.Message = "empty package"
		}
		if err == nil {
			result, err = json.Marshal(response)
			if err != nil {
				ylog.Error(err)
			}
			if err := conn.SendPkg(result); err != nil {
				ylog.Error(err)
			}
		} else {
			// Just close the connection if any error occurs.
			if err := conn.Close(); err != nil {
				ylog.Error(err)
			}
			break
		}
	}
}
