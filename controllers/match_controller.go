package controllers

import (
	"net/http"

	"github.com/bitspawngg/tournament-bracket-manager/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MatchController struct {
	log *logrus.Entry
	ms  *services.MatchService
}

func NewMatchController(log *logrus.Logger, ms *services.MatchService) *MatchController {
	return &MatchController{
		log: log.WithField("controller", "match"),
		ms:  ms,
	}
}

func (mc *MatchController) HandlePing(c *gin.Context) {
	mc.log.Info("handling ping")
	c.JSON(
		http.StatusOK,
		gin.H{
			"msg": "pong",
		},
	)
}

type FormGetMatchSchedule struct {
	Teams  []string `json:"teams"`
	Format string   `json:"format"`
}

type FormSetMatchResult struct {
	Tournament string `json:"tournament"`
	Round      int    `json:"round"`
	Table      int    `json:"table"`
	Result     int    `json:"result"`
}

type FormGetMatches struct {
	Tournament string `json:"tournament"`
}

func (mc *MatchController) HandleGetMatchSchedule(c *gin.Context) {
	form := FormGetMatchSchedule{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"msg":   "failure",
				"error": err.Error(),
			},
		)
		return
	}

	if form.Teams == nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"msg":   "failure",
				"error": "missing mandatory input parameter",
			},
		)
		return
	}

	brackets, id, err := services.GetMatchSchedule(form.Teams, form.Format)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"msg":   "failure",
				"error": err.Error(),
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"msg":           "success",
			"data":          brackets,
			"tournament_id": id,
		},
	)
}

func (mc *MatchController) HandleSetMatchResult(c *gin.Context) {
	form := FormSetMatchResult{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"msg":   "failure",
				"error": err.Error(),
			},
		)
		return
	}

	if form.Tournament == "" || form.Round == 0 || form.Table == 0 {
		c.JSON(http.StatusBadRequest,
			gin.H{
				"msg":   "failure",
				"error": "missing param",
			},
		)
		return
	}

	if err := services.SetMatchResult(form.Tournament, form.Round, form.Table, form.Result); err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"msg":   "failure",
				"error": err.Error(),
			},
		)
		return
	}

	c.JSON(http.StatusOK,
		gin.H{
			"msg": "success",
		},
	)
}

func (mc *MatchController) HandleGetMatches(c *gin.Context) {
	form := FormGetMatches{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"msg": "failure",
				"err": err.Error(),
			},
		)
		return
	}

	if form.Tournament == "" {
		c.JSON(http.StatusBadRequest,
			gin.H{
				"msg": "failure",
				"err": "tournament required",
			},
		)
		return
	}

	matches, err := services.GetMatches(form.Tournament)
	if err != nil {
		mc.log.Error(err.Error())
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"msg": "failure",
				"err": err.Error(),
			},
		)
		return
	}

	c.JSON(http.StatusOK,
		gin.H{
			"msg":  "success",
			"data": matches,
		},
	)
}
