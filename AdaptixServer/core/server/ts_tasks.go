package server

import (
	"AdaptixServer/core/adaptix"
	"AdaptixServer/core/utils/krypt"
	"AdaptixServer/core/utils/logs"
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

func (ts *Teamserver) TsTaskQueueAddQuite(agentId string, taskObject []byte) {
	var (
		agent    *Agent
		taskData adaptix.TaskData
		value    any
		ok       bool
		err      error
	)
	err = json.Unmarshal(taskObject, &taskData)
	if err != nil {
		return
	}

	taskData.Sync = false

	value, ok = ts.agents.Get(agentId)
	if ok {
		agent = value.(*Agent)
	} else {
		logs.Error("", "TsTaskQueueAdd: agent %v not found", agentId)
		return
	}

	agent.TasksQueue.Put(taskData)
}

func (ts *Teamserver) TsTaskQueueGetAvailable(agentId string, availableSize int) ([][]byte, error) {
	var (
		tasksArray [][]byte
		agent      *Agent
		value      any
		ok         bool
	)

	value, ok = ts.agents.Get(agentId)
	if ok {
		agent = value.(*Agent)
	} else {
		return nil, fmt.Errorf("TsTaskQueueGetAvailable: agent %v not found", agentId)
	}

	/// TASKS QUEUE

	for i := 0; i < agent.TasksQueue.Len(); i++ {
		value, ok = agent.TasksQueue.Get(i)
		if ok {
			taskData := value.(adaptix.TaskData)
			if len(tasksArray)+len(taskData.Data) < availableSize {
				var taskBuffer bytes.Buffer
				_ = json.NewEncoder(&taskBuffer).Encode(taskData)
				tasksArray = append(tasksArray, taskBuffer.Bytes())
				agent.Tasks.Put(taskData.TaskId, taskData)
				agent.TasksQueue.Delete(i)
				i--
			} else {
				break
			}
		} else {
			break
		}
	}

	/// TUNNELS QUEUE

	for i := 0; i < agent.TunnelQueue.Len(); i++ {
		value, ok = agent.TunnelQueue.Get(i)
		if ok {
			tunnelTaskData := value.(adaptix.TaskData)
			if len(tasksArray)+len(tunnelTaskData.Data) < availableSize {
				var taskBuffer bytes.Buffer
				_ = json.NewEncoder(&taskBuffer).Encode(tunnelTaskData)
				tasksArray = append(tasksArray, taskBuffer.Bytes())
				agent.TunnelQueue.Delete(i)
				i--
			} else {
				break
			}
		} else {
			break
		}
	}

	return tasksArray, nil
}

func (ts *Teamserver) TsTaskUpdate(agentId string, taskObject []byte) {
	var (
		agent    *Agent
		task     adaptix.TaskData
		taskData adaptix.TaskData
		value    any
		ok       bool
		err      error
	)
	err = json.Unmarshal(taskObject, &taskData)
	if err != nil {
		return
	}

	value, ok = ts.agents.Get(agentId)
	if ok {
		agent = value.(*Agent)
	} else {
		logs.Error("", "TsTaskUpdate: agent %v not found", agentId)
		return
	}

	value, ok = agent.Tasks.GetDelete(taskData.TaskId)
	if !ok {
		return
	}

	task = value.(adaptix.TaskData)
	task.Data = []byte("")
	task.FinishDate = taskData.FinishDate
	task.Completed = taskData.Completed

	if task.Type == TYPE_JOB {
		if task.MessageType != 6 {
			task.MessageType = taskData.MessageType
		}

		var oldMessage string
		if task.Message == "" {
			oldMessage = taskData.Message
		} else {
			oldMessage = task.Message
		}

		oldText := task.ClearText

		task.Message = taskData.Message
		task.ClearText = taskData.ClearText

		packet := CreateSpAgentTaskUpdate(task)

		task.Message = oldMessage
		task.ClearText = oldText + task.ClearText

		if task.Completed {
			agent.ClosedTasks.Put(task.TaskId, task)
		} else {
			agent.Tasks.Put(task.TaskId, task)
		}

		if task.Sync {
			if task.Completed {
				ts.DBMS.DbTaskInsert(task)
			}

			ts.TsSyncAllClients(packet)
		}

	} else if task.Type == TYPE_TUNNEL {
		if task.Completed {
			task.Message = taskData.Message
			task.MessageType = taskData.MessageType
			
			agent.ClosedTasks.Put(task.TaskId, task)

			if task.Sync {
				ts.DBMS.DbTaskInsert(task)
				packet := CreateSpAgentTaskUpdate(task)
				ts.TsSyncAllClients(packet)
			}
		}

	} else if task.Type == TYPE_TASK || task.Type == TYPE_BROWSER {
		task.MessageType = taskData.MessageType
		task.Message = taskData.Message
		task.ClearText = taskData.ClearText

		if task.Completed {
			agent.ClosedTasks.Put(task.TaskId, task)
		} else {
			agent.Tasks.Put(task.TaskId, task)
		}

		if task.Sync {
			if task.Completed {
				ts.DBMS.DbTaskInsert(task)
			}

			packet := CreateSpAgentTaskUpdate(task)
			ts.TsSyncAllClients(packet)
		}
	}
}

func (ts *Teamserver) TsTaskStop(agentId string, taskId string) error {
	var (
		agent    *Agent
		task     adaptix.TaskData
		taskData adaptix.TaskData
		value    any
		ok       bool
		found    bool
	)

	value, ok = ts.agents.Get(agentId)
	if ok {
		agent = value.(*Agent)
	} else {
		return fmt.Errorf("agent %v not found", agentId)
	}

	found = false
	for i := 0; i < agent.TasksQueue.Len(); i++ {
		if value, ok = agent.TasksQueue.Get(i); ok {
			task = value.(adaptix.TaskData)
			if task.TaskId == taskId {
				agent.TasksQueue.Delete(i)
				found = true
				break
			}
		}
	}

	if found {
		packet := CreateSpAgentTaskRemove(task)
		ts.TsSyncAllClients(packet)
		return nil
	}

	value, ok = agent.Tasks.Get(taskId)
	if ok {
		task = value.(adaptix.TaskData)
		if task.Type == TYPE_JOB {
			data, err := ts.Extender.ExAgentBrowserJobKill(agent.Data.Name, taskId)
			if err != nil {
				return err
			}

			err = json.Unmarshal(data, &taskData)
			if err != nil {
				return err
			}

			if taskData.TaskId == "" {
				taskData.TaskId, _ = krypt.GenerateUID(8)
			}
			taskData.AgentId = agentId
			taskData.CommandLine = "job kill " + taskId
			taskData.StartDate = time.Now().Unix()
			agent.TasksQueue.Put(taskData)
			return nil
		} else {
			return fmt.Errorf("taski %v in process", taskId)
		}
	}
	return nil
}

func (ts *Teamserver) TsTaskDelete(agentId string, taskId string) error {
	var (
		agent *Agent
		task  adaptix.TaskData
		value any
		ok    bool
	)

	value, ok = ts.agents.Get(agentId)
	if ok {
		agent = value.(*Agent)
	} else {
		return fmt.Errorf("agent %v not found", agentId)
	}

	for i := 0; i < agent.TasksQueue.Len(); i++ {
		if value, ok = agent.TasksQueue.Get(i); ok {
			task = value.(adaptix.TaskData)
			if task.TaskId == taskId {
				return fmt.Errorf("task %v in process", taskId)
			}
		}
	}

	value, ok = agent.Tasks.Get(taskId)
	if ok {
		return fmt.Errorf("task %v in process", taskId)
	}

	value, ok = agent.ClosedTasks.GetDelete(taskId)
	if ok {
		task = value.(adaptix.TaskData)
		ts.DBMS.DbTaskDelete(task.TaskId, "")

		packet := CreateSpAgentTaskRemove(task)
		ts.TsSyncAllClients(packet)
		return nil
	}

	return fmt.Errorf("task %v not found", taskId)
}
