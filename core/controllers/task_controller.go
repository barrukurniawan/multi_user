package controllers

import (
	"net/http"

	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/barrukurniawan/multi_user/core/helpers"
	"github.com/barrukurniawan/multi_user/core/services"
	"github.com/gin-gonic/gin"
)

type TaskController struct {
	taskService services.TaskService
}

func NewTaskController(taskService services.TaskService) *TaskController {
	return &TaskController{taskService: taskService}
}

func (ctl *TaskController) Create(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.WriteBindError(c, err)
		return
	}

	key := c.GetHeader("Idempotency-Key")
	task, statusCode, err := ctl.taskService.Create(c.Request.Context(), helpers.CurrentUserID(c), key, req)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}
	c.JSON(statusCode, dto.Success(task))
}

func (ctl *TaskController) List(c *gin.Context) {
	var query dto.ListTaskQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		helpers.WriteBindError(c, err)
		return
	}

	tasks, meta, err := ctl.taskService.List(c.Request.Context(), helpers.CurrentUserID(c), query)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.List(tasks, meta))
}

func (ctl *TaskController) GetByID(c *gin.Context) {
	id, ok := helpers.TaskIDParam(c)
	if !ok {
		return
	}

	task, err := ctl.taskService.GetByID(c.Request.Context(), helpers.CurrentUserID(c), id)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Success(task))
}

func (ctl *TaskController) Update(c *gin.Context) {
	id, ok := helpers.TaskIDParam(c)
	if !ok {
		return
	}

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.WriteBindError(c, err)
		return
	}

	task, err := ctl.taskService.Update(c.Request.Context(), helpers.CurrentUserID(c), id, req)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Success(task))
}

func (ctl *TaskController) Delete(c *gin.Context) {
	id, ok := helpers.TaskIDParam(c)
	if !ok {
		return
	}

	if err := ctl.taskService.Delete(c.Request.Context(), helpers.CurrentUserID(c), id); err != nil {
		helpers.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Success(gin.H{"message": "task berhasil dihapus"}))
}

func (ctl *TaskController) Assign(c *gin.Context) {
	id, ok := helpers.TaskIDParam(c)
	if !ok {
		return
	}

	var req dto.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.WriteBindError(c, err)
		return
	}

	task, err := ctl.taskService.Assign(c.Request.Context(), helpers.CurrentUserID(c), id, req.AssigneeID)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Success(task))
}
