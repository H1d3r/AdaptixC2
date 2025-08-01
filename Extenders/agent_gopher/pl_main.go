package main

import (
	"encoding/hex"
	"encoding/json"
	"github.com/Adaptix-Framework/axc2"
	"math/rand"
	"time"
)

const (
	OS_UNKNOWN = 0
	OS_WINDOWS = 1
	OS_LINUX   = 2
	OS_MAC     = 3

	TYPE_TASK       = 1
	TYPE_BROWSER    = 2
	TYPE_JOB        = 3
	TYPE_TUNNEL     = 4
	TYPE_PROXY_DATA = 5

	MESSAGE_INFO    = 5
	MESSAGE_ERROR   = 6
	MESSAGE_SUCCESS = 7
)

type Teamserver interface {
	TsListenerInteralHandler(watermark string, data []byte) (string, error)

	TsAgentProcessData(agentId string, bodyData []byte) error

	TsAgentUpdateData(newAgentData adaptix.AgentData) error
	TsAgentTerminate(agentId string, terminateTaskId string) error
	TsAgentSetImpersonate(agentId string, impersonated string, elevated bool) error

	TsAgentConsoleOutput(agentId string, messageType int, message string, clearText string, store bool)
	TsAgentConsoleOutputClient(agentId string, client string, messageType int, message string, clearText string)

	TsPivotCreate(pivotId string, pAgentId string, chAgentId string, pivotName string, isRestore bool) error
	TsGetPivotInfoByName(pivotName string) (string, string, string)
	TsGetPivotInfoById(pivotId string) (string, string, string)
	TsPivotDelete(pivotId string) error

	TsTaskCreate(agentId string, cmdline string, client string, taskData adaptix.TaskData)
	TsTaskUpdate(agentId string, data adaptix.TaskData)
	TsTaskGetAvailableAll(agentId string, availableSize int) ([]adaptix.TaskData, error)

	TsDownloadAdd(agentId string, fileId string, fileName string, fileSize int) error
	TsDownloadUpdate(fileId string, state int, data []byte) error
	TsDownloadClose(fileId string, reason int) error

	TsScreenshotAdd(agentId string, Note string, Content []byte) error

	TsClientGuiDisks(taskData adaptix.TaskData, jsonDrives string)
	TsClientGuiFiles(taskData adaptix.TaskData, path string, jsonFiles string)
	TsClientGuiFilesStatus(taskData adaptix.TaskData)
	TsClientGuiProcess(taskData adaptix.TaskData, jsonFiles string)

	TsTunnelStart(TunnelId string) (string, error)
	TsTunnelCreateSocks4(AgentId string, Info string, Lhost string, Lport int) (string, error)
	TsTunnelCreateSocks5(AgentId string, Info string, Lhost string, Lport int, UseAuth bool, Username string, Password string) (string, error)
	TsTunnelCreateLportfwd(AgentId string, Info string, Lhost string, Lport int, Thost string, Tport int) (string, error)
	TsTunnelCreateRportfwd(AgentId string, Info string, Lport int, Thost string, Tport int) (string, error)
	TsTunnelUpdateRportfwd(tunnelId int, result bool) (string, string, error)

	TsTunnelStopSocks(AgentId string, Port int)
	TsTunnelStopLportfwd(AgentId string, Port int)
	TsTunnelStopRportfwd(AgentId string, Port int)

	TsTunnelConnectionClose(channelId int)
	TsTunnelConnectionResume(AgentId string, channelId int, ioDirect bool)
	TsTunnelConnectionData(channelId int, data []byte)
	TsTunnelConnectionAccept(tunnelId int, channelId int)
}

type ModuleExtender struct {
	ts Teamserver
}

var (
	ModuleObject   *ModuleExtender
	ModuleDir      string
	AgentWatermark string
)

func InitPlugin(ts any, moduleDir string, watermark string) any {
	ModuleDir = moduleDir
	AgentWatermark = watermark

	ModuleObject = &ModuleExtender{
		ts: ts.(Teamserver),
	}
	return ModuleObject
}

func (m *ModuleExtender) AgentGenerate(config string, listenerWM string, listenerProfile []byte) ([]byte, string, error) {
	var (
		listenerMap  map[string]any
		agentProfile []byte
		err          error
	)

	err = json.Unmarshal(listenerProfile, &listenerMap)
	if err != nil {
		return nil, "", err
	}

	agentProfile, err = AgentGenerateProfile(config, listenerWM, listenerMap)
	if err != nil {
		return nil, "", err
	}

	return AgentGenerateBuild(config, agentProfile, listenerMap)
}

func (m *ModuleExtender) AgentCreate(beat []byte) (adaptix.AgentData, error) {
	return CreateAgent(beat)
}

func (m *ModuleExtender) AgentCommand(agentData adaptix.AgentData, args map[string]any) (adaptix.TaskData, adaptix.ConsoleMessageData, error) {
	return CreateTask(m.ts, agentData, args)
}

func (m *ModuleExtender) AgentPackData(agentData adaptix.AgentData, tasks []adaptix.TaskData) ([]byte, error) {
	packedData, err := PackTasks(agentData, tasks)
	if err != nil {
		return nil, err
	}

	return AgentEncryptData(packedData, agentData.SessionKey)
}

func (m *ModuleExtender) AgentPivotPackData(pivotId string, data []byte) (adaptix.TaskData, error) {
	packData, err := PackPivotTasks(pivotId, data)
	if err != nil {
		return adaptix.TaskData{}, err
	}

	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	uid := hex.EncodeToString(randomBytes)[:8]

	taskData := adaptix.TaskData{
		TaskId: uid,
		Type:   TYPE_PROXY_DATA,
		Data:   packData,
		Sync:   false,
	}

	return taskData, nil
}

func (m *ModuleExtender) AgentProcessData(agentData adaptix.AgentData, packedData []byte) ([]byte, error) {
	decryptData, err := AgentDecryptData(packedData, agentData.SessionKey)
	if err != nil {
		return nil, err
	}

	taskData := adaptix.TaskData{
		Type:        TYPE_TASK,
		AgentId:     agentData.Id,
		FinishDate:  time.Now().Unix(),
		MessageType: MESSAGE_SUCCESS,
		Completed:   true,
		Sync:        true,
	}

	resultTasks := ProcessTasksResult(m.ts, agentData, taskData, decryptData)

	for _, task := range resultTasks {
		m.ts.TsTaskUpdate(agentData.Id, task)
	}

	return nil, nil
}

/// SYNC

func SyncBrowserDisks(ts Teamserver, taskData adaptix.TaskData, drivesSlice []adaptix.ListingDrivesDataWin) {
	jsonDrives, err := json.Marshal(drivesSlice)
	if err != nil {
		return
	}

	ts.TsClientGuiDisks(taskData, string(jsonDrives))
}

func SyncBrowserFilesUnix(ts Teamserver, taskData adaptix.TaskData, path string, filesSlice []adaptix.ListingFileDataUnix) {
	jsonDrives, err := json.Marshal(filesSlice)
	if err != nil {
		return
	}

	ts.TsClientGuiFiles(taskData, path, string(jsonDrives))
}

func SyncBrowserFilesWindows(ts Teamserver, taskData adaptix.TaskData, path string, filesSlice []adaptix.ListingFileDataWin) {
	jsonDrives, err := json.Marshal(filesSlice)
	if err != nil {
		return
	}

	ts.TsClientGuiFiles(taskData, path, string(jsonDrives))
}

func SyncBrowserFilesStatus(ts Teamserver, taskData adaptix.TaskData) {
	ts.TsClientGuiFilesStatus(taskData)
}

func SyncBrowserProcessUnix(ts Teamserver, taskData adaptix.TaskData, processlist []adaptix.ListingProcessDataUnix) {
	jsonProcess, err := json.Marshal(processlist)
	if err != nil {
		return
	}

	ts.TsClientGuiProcess(taskData, string(jsonProcess))
}

func SyncBrowserProcessWindows(ts Teamserver, taskData adaptix.TaskData, processlist []adaptix.ListingProcessDataWin) {
	jsonProcess, err := json.Marshal(processlist)
	if err != nil {
		return
	}

	ts.TsClientGuiProcess(taskData, string(jsonProcess))
}

/// TUNNEL

func (m *ModuleExtender) AgentTunnelCallbacks() (func(channelId int, address string, port int) adaptix.TaskData, func(channelId int, address string, port int) adaptix.TaskData, func(channelId int, data []byte) adaptix.TaskData, func(channelId int, data []byte) adaptix.TaskData, func(channelId int) adaptix.TaskData, func(tunnelId int, port int) adaptix.TaskData, error) {
	return TunnelMessageConnectTCP, TunnelMessageConnectUDP, TunnelMessageWriteTCP, TunnelMessageWriteUDP, TunnelMessageClose, TunnelMessageReverse, nil
}

func TunnelMessageConnectTCP(channelId int, address string, port int) adaptix.TaskData {
	packData, _ := TunnelCreateTCP(channelId, address, port)

	taskData := adaptix.TaskData{
		Type: TYPE_PROXY_DATA,
		Data: packData,
		Sync: false,
	}

	return taskData
}

func TunnelMessageConnectUDP(channelId int, address string, port int) adaptix.TaskData {
	packData, _ := TunnelCreateUDP(channelId, address, port)

	taskData := adaptix.TaskData{
		Type: TYPE_PROXY_DATA,
		Data: packData,
		Sync: false,
	}

	return taskData
}

func TunnelMessageWriteTCP(channelId int, data []byte) adaptix.TaskData {
	packData, _ := TunnelWriteTCP(channelId, data)

	taskData := adaptix.TaskData{
		Type: TYPE_PROXY_DATA,
		Data: packData,
		Sync: false,
	}

	return taskData
}

func TunnelMessageWriteUDP(channelId int, data []byte) adaptix.TaskData {
	packData, _ := TunnelWriteUDP(channelId, data)

	taskData := adaptix.TaskData{
		Type: TYPE_PROXY_DATA,
		Data: packData,
		Sync: false,
	}

	return taskData
}

func TunnelMessageClose(channelId int) adaptix.TaskData {
	packData, _ := TunnelClose(channelId)

	taskData := adaptix.TaskData{
		Type: TYPE_PROXY_DATA,
		Data: packData,
		Sync: false,
	}

	return taskData
}

func TunnelMessageReverse(tunnelId int, port int) adaptix.TaskData {
	packData, _ := TunnelReverse(tunnelId, port)

	taskData := adaptix.TaskData{
		Type: TYPE_PROXY_DATA,
		Data: packData,
		Sync: false,
	}

	return taskData
}

/// TERMINAL

func (m *ModuleExtender) AgentTerminalCallbacks() (func(int, string, int, int) (adaptix.TaskData, error), func(int, []byte) (adaptix.TaskData, error), func(int) (adaptix.TaskData, error), error) {
	return TerminalMessageStart, TerminalMessageWrite, TerminalMessageClose, nil
}

func TerminalMessageStart(terminalId int, program string, sizeH int, sizeW int) (adaptix.TaskData, error) {
	packData, err := TerminalStart(terminalId, program, sizeH, sizeW)
	if err != nil {
		return adaptix.TaskData{}, err
	}

	taskData := adaptix.TaskData{
		Type: TYPE_PROXY_DATA,
		Data: packData,
		Sync: false,
	}

	return taskData, nil
}

func TerminalMessageWrite(channelId int, data []byte) (adaptix.TaskData, error) {
	packData, err := TerminalWrite(channelId, data)
	if err != nil {
		return adaptix.TaskData{}, err
	}
	taskData := adaptix.TaskData{
		Type: TYPE_PROXY_DATA,
		Data: packData,
		Sync: false,
	}

	return taskData, nil
}

func TerminalMessageClose(terminalId int) (adaptix.TaskData, error) {
	packData, err := TerminalClose(terminalId)
	if err != nil {
		return adaptix.TaskData{}, err
	}

	taskData := adaptix.TaskData{
		Type: TYPE_PROXY_DATA,
		Data: packData,
		Sync: false,
	}

	return taskData, nil
}
