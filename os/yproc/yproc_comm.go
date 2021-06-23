package yproc

import (
	"errors"
	"fmt"

	"github.com/AmarsDing/lib/container/ymap"
	"github.com/AmarsDing/lib/net/ytcp"
	"github.com/AmarsDing/lib/os/yfile"
	"github.com/AmarsDing/lib/util/yconv"
)

// MsgRequest is the request structure for process communication.
type MsgRequest struct {
	SendPid int    // Sender PID.
	RecvPid int    // Receiver PID.
	Group   string // Message group name.
	Data    []byte // Request data.
}

// MsgResponse is the response structure for process communication.
type MsgResponse struct {
	Code    int    // 1: OK; Other: Error.
	Message string // Response message.
	Data    []byte // Response data.
}

const (
	gPROC_COMM_DEFAULT_GRUOP_NAME = ""    // Default group name.
	gPROC_DEFAULT_TCP_PORT        = 10000 // Starting port number for receiver listening.
	gPROC_MSG_QUEUE_MAX_LENGTH    = 10000 // Max size for each message queue of the group.
)

var (
	// commReceiveQueues is the group name to queue map for storing received data.
	// The value of the map is type of *gqueue.Queue.
	commReceiveQueues = ymap.NewStrAnyMap(true)

	// commPidFolderPath specifies the folder path storing pid to port mapping files.
	commPidFolderPath = yfile.TempDir("gproc")
)

func init() {
	// Automatically create the storage folder.
	if !yfile.Exists(commPidFolderPath) {
		err := yfile.Mkdir(commPidFolderPath)
		if err != nil {
			panic(fmt.Errorf(`create gproc folder failed: %v`, err))
		}
	}
}

// getConnByPid creates and returns a TCP connection for specified pid.
func getConnByPid(pid int) (*ytcp.PoolConn, error) {
	port := getPortByPid(pid)
	if port > 0 {
		if conn, err := ytcp.NewPoolConn(fmt.Sprintf("127.0.0.1:%d", port)); err == nil {
			return conn, nil
		} else {
			return nil, err
		}
	}
	return nil, errors.New(fmt.Sprintf("could not find port for pid: %d", pid))
}

// getPortByPid returns the listening port for specified pid.
// It returns 0 if no port found for the specified pid.
func getPortByPid(pid int) int {
	path := getCommFilePath(pid)
	content := yfile.GetContentsWithCache(path)
	return yconv.Int(content)
}

// getCommFilePath returns the pid to port mapping file path for given pid.
func getCommFilePath(pid int) string {
	return yfile.Join(commPidFolderPath, yconv.String(pid))
}
