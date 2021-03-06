package gae

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/taskqueue"
	"net/url"
	"time"
)

//TODO: Document why whe need this
var CallDelayFunc = func(c context.Context, queueName, subPath string, f *delay.Function, args ...interface{}) error {
	return CallDelayFuncWithDelay(c, 0, queueName, subPath, f, args...)
}

var CallDelayFuncWithDelay = func(c context.Context, delay time.Duration, queueName, subPath string, f *delay.Function, args ...interface{}) error {
	if task, err := CreateDelayTask(queueName, subPath, f, args...); err != nil {
		return err
	} else {
		task.Delay = delay
		_, err = AddTaskToQueue(c, task, queueName)
		return err
	}
}

const failToCreateDelayTask = "failed to create delay task"
const failToCreateDelayTaskPrefix = failToCreateDelayTask + ": "

//TODO: Document why whe need this
func CreateDelayTask(queueName, subPath string, f *delay.Function, args ...interface{}) (*taskqueue.Task, error) {
	if queueName == "" {
		return nil, errors.New(failToCreateDelayTaskPrefix + "queueName is empty")
	}
	if queueName == "default" {
		return nil, errors.New(failToCreateDelayTaskPrefix + "queueName is 'default'")
	}
	if subPath == "" {
		return nil, errors.New(failToCreateDelayTaskPrefix + "subPath is empty")
	}
	if task, err := f.Task(args...); err != nil {
		err = errors.WithMessage(err, fmt.Sprintf("queue=%v, subPath=%v", queueName, subPath))
		return task, errors.Wrap(err, failToCreateDelayTask) // Use wrap intentionally
	} else {
		task.Path += fmt.Sprintf("?task=%v&queue=%v", url.QueryEscape(subPath), url.QueryEscape(queueName))
		return task, nil
	}
}

const failedToAddTaskToQueue = "faile to add task to queue"
const failedToAddTaskToQueuePrefix = failedToAddTaskToQueue + ": "

//TODO: Document why whe need this
var AddTaskToQueue = func(c context.Context, t *taskqueue.Task, queueName string) (task *taskqueue.Task, err error) {
	if queueName == "" {
		return nil, errors.New(failedToAddTaskToQueuePrefix + "queueName is empty")
	}
	if queueName == "default" {
		return nil, errors.New(failedToAddTaskToQueuePrefix + "queueName is 'default'")
	}
	if task, err = taskqueue.Add(c, t, queueName); err != nil {
		err = errors.WithMessage(err, failedToAddTaskToQueue)
	} else {
		isInTransaction := gaedb.NewDatabase().IsInTransaction(c)
		log.Debugf(c, "Added task to queue '%v', tx=%v): path: %v", queueName, isInTransaction, task.Path)
	}
	return
}
