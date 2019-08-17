/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package todotxt

import (
//	"fmt"
  "gopkg.in/yaml.v2"
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

//TODO: do this again with a raw text string whenactually parsing a CE entry block.
// transform task from yaml object to a task
//func loadCETask(ceTask 
// ParseTask parses the input text string into a Task struct.
func ParseCETask(text string) (*Task, error) {
	var err error

	task := Task{}
	task.Original = strings.Trim(text, "\t\n\r ")
	task.Todo = task.Original

	// Check for completed
	if completedRx.MatchString(task.Original) {
		task.Completed = true
		// Check for completed date
		if completedDateRx.MatchString(task.Original) {
			if date, err := time.Parse(DateLayout, completedDateRx.FindStringSubmatch(task.Original)[1]); err == nil {
				task.CompletedDate = date
			} else {
				return nil, err
			}
		}

		// Remove from Todo text
		task.Todo = completedDateRx.ReplaceAllString(task.Todo, "") // Strip CompletedDate first, otherwise it wouldn't match anymore (^x date...)
		task.Todo = completedRx.ReplaceAllString(task.Todo, "")     // Strip 'x '
	}

	// Check for priority
	if priorityRx.MatchString(task.Original) {
		task.Priority = priorityRx.FindStringSubmatch(task.Original)[2]
		task.Todo = priorityRx.ReplaceAllString(task.Todo, "") // Remove from Todo text
	}

	// Check for created date
	if createdDateRx.MatchString(task.Original) {
		if date, err := time.Parse(DateLayout, createdDateRx.FindStringSubmatch(task.Original)[2]); err == nil {
			task.CreatedDate = date
			task.Todo = createdDateRx.ReplaceAllString(task.Todo, "") // Remove from Todo text
		} else {
			return nil, err
		}
	}

	// function for collecting projects/contexts as slices from text
	getSlice := func(rx *regexp.Regexp) []string {
		matches := rx.FindAllStringSubmatch(task.Original, -1)
		slice := make([]string, 0, len(matches))
		seen := make(map[string]bool, len(matches))
		for _, match := range matches {
			word := strings.Trim(match[2], "\t\n\r ")
			if _, found := seen[word]; !found {
				slice = append(slice, word)
				seen[word] = true
			}
		}
		sort.Strings(slice)
		return slice
	}

	// Check for contexts
	if contextRx.MatchString(task.Original) {
		task.Contexts = getSlice(contextRx)
		task.Todo = contextRx.ReplaceAllString(task.Todo, "") // Remove from Todo text
	}

	// Check for projects
	if projectRx.MatchString(task.Original) {
		task.Projects = getSlice(projectRx)
		task.Todo = projectRx.ReplaceAllString(task.Todo, "") // Remove from Todo text
	}

	// Check for additional tags
	if addonTagRx.MatchString(task.Original) {
		matches := addonTagRx.FindAllStringSubmatch(task.Original, -1)
		tags := make(map[string]string, len(matches))
		for _, match := range matches {
			key, value := match[2], match[3]
			if key == "due" { // due date is a known addon tag, it has its own struct field
				if date, err := time.Parse(DateLayout, value); err == nil {
					task.DueDate = date
				} else {
					return nil, err
				}
			} else if key != "" && value != "" {
				tags[key] = value
			}
		}
		task.AdditionalTags = tags
		task.Todo = addonTagRx.ReplaceAllString(task.Todo, "") // Remove from Todo text
	}

	// Trim any remaining whitespaces from Todo text
	//TODO: add <br> where newlines are.
	task.Todo = strings.Trim(task.Todo, "\t\n\r\f ")

	return &task, err
}

// See *Task.String() for further information.
// in task.go for taking tasks and printing in CE format (not sure why that's needed at this moment though...


// TaskList represents a list of todo.txt task entries.
// It is usually loaded from a whole todo.txt file.
type CETaskList []CETask

type CETask struct {
		Original       string `json:"content"`// Original raw task text.
		Priority       string `json:"priority"`
		Projects       []string `json:"projects"`
		Contexts       []string `json:"contexts"`
    AdditionalTags map[string]string // Addon tags will be available here.
		CreatedDate    time.Time `json:"createdDate"`
		CompletedDate  time.Time `json:"completedDate"`
		Completed      bool `json:"completed"`
}

// LoadFromCEFile loads a CETaskList yaml file from *os.File.
//
// Using *os.File instead of a filename allows to also use os.Stdin.
//
// Note: This will clear the current TaskList and overwrite it's contents with whatever is in *os.File.
func (tasklist *TaskList) LoadFromCEFile(file *os.File) error {
	*tasklist = []Task{} // Empty tasklist

	taskId := 1

	yamlFile, err := ioutil.ReadFile(file)
    if err != nil {
        log.Printf("yamlFile.Get err   #%v ", err)
    }

		err = yaml.Unmarshal(yamlFile, taskList)
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }
  //parse yaml file
	//loop through each
	//parse them is really translate them
	{
		task, err := ParseCETask(taskYML)
		if err != nil {
			return err
		}
		task.Id = taskId

		*tasklist = append(*tasklist, *task)
		taskId++
  //}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
