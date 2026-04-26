// AstrCode - Agent orchestration engine for AstrBot
// Copyright (C) 2026 EterUltimate
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package model

import (
	"testing"
	"time"
)

func TestTaskStoreCRUD(t *testing.T) {
	store := NewTaskStore()

	// Create
	task := &Task{
		ID:        "task_test_1",
		Content:   "test task",
		Status:    TaskStatusPending,
		CreatedAt: time.Now().Unix(),
	}
	store.CreateTask(task)

	// Get
	got, ok := store.GetTask("task_test_1")
	if !ok {
		t.Fatal("task not found")
	}
	if got.Content != "test task" {
		t.Errorf("expected 'test task', got '%s'", got.Content)
	}

	// Update
	got.Status = TaskStatusCompleted
	got.CompletedAt = time.Now().Unix()
	store.UpdateTask(got)

	updated, _ := store.GetTask("task_test_1")
	if updated.Status != TaskStatusCompleted {
		t.Errorf("expected completed, got %s", updated.Status)
	}

	// List
	tasks := store.ListTasks(10)
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	// ListRunning (none)
	running := store.ListRunning()
	if len(running) != 0 {
		t.Errorf("expected 0 running, got %d", len(running))
	}
}

func TestTaskStoreRunning(t *testing.T) {
	store := NewTaskStore()
	store.CreateTask(&Task{ID: "t1", Status: TaskStatusExecuting})
	store.CreateTask(&Task{ID: "t2", Status: TaskStatusPlanning})
	store.CreateTask(&Task{ID: "t3", Status: TaskStatusCompleted})

	running := store.ListRunning()
	if len(running) != 2 {
		t.Errorf("expected 2 running, got %d", len(running))
	}
}

func TestPlanGraph(t *testing.T) {
	plan := &Plan{
		ID: "plan_test",
		Steps: []Step{
			{ID: "s1", Type: StepTypeSkill, Skill: "skill_a", Status: StepStatusPending},
			{ID: "s2", Type: StepTypeSkill, Skill: "skill_b", Status: StepStatusPending, DependsOn: []string{"s1"}},
			{ID: "s3", Type: StepTypeSkill, Skill: "skill_c", Status: StepStatusPending, DependsOn: []string{"s1"}},
			{ID: "s4", Type: StepTypeSkill, Skill: "skill_d", Status: StepStatusPending, DependsOn: []string{"s2", "s3"}},
		},
	}

	graph := plan.BuildGraph()

	// Initially only s1 is ready
	ready := graph.GetReadySteps()
	if len(ready) != 1 || ready[0].ID != "s1" {
		t.Errorf("expected s1 ready, got %v", ready)
	}

	// Complete s1
	graph.Nodes["s1"].Status = StepStatusCompleted
	ready = graph.GetReadySteps()
	if len(ready) != 2 {
		t.Errorf("expected 2 ready after s1, got %d", len(ready))
	}

	// Complete s2, s3
	graph.Nodes["s2"].Status = StepStatusCompleted
	graph.Nodes["s3"].Status = StepStatusCompleted
	ready = graph.GetReadySteps()
	if len(ready) != 1 || ready[0].ID != "s4" {
		t.Errorf("expected s4 ready, got %v", ready)
	}

	// Complete s4
	graph.Nodes["s4"].Status = StepStatusCompleted
	if !graph.IsComplete() {
		t.Error("graph should be complete")
	}
}

func TestStepTimeline(t *testing.T) {
	timeline := &StepTimeline{}
	timeline.Add("s1", "running", "start s1")
	timeline.AddWithDuration("s1", "completed", "done", time.Now().UnixNano()-1000000)

	if len(timeline.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(timeline.Items))
	}
	if timeline.Items[0].StepID != "s1" {
		t.Error("wrong step ID")
	}
}

func TestBuildSnapshot(t *testing.T) {
	task := &Task{ID: "t1", Status: TaskStatusCompleted, CreatedAt: 1000, CompletedAt: 2000}
	plan := &Plan{
		ID: "p1",
		Steps: []Step{
			{ID: "s1", Type: StepTypeSkill, Skill: "test", Status: StepStatusCompleted},
		},
	}

	snap := BuildSnapshot(task, plan, nil)
	if snap.TaskID != "t1" {
		t.Error("wrong task ID")
	}
	if len(snap.Steps) != 1 {
		t.Error("wrong step count")
	}
}

func TestWSEvent(t *testing.T) {
	evt := NewWSEvent("step_started", "t1", "s1", map[string]string{"skill": "test"})
	if evt.Type != "step_started" {
		t.Error("wrong event type")
	}
	if evt.TaskID != "t1" {
		t.Error("wrong task ID")
	}
	if evt.StepID != "s1" {
		t.Error("wrong step ID")
	}
}
