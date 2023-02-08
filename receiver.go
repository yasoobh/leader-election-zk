package main

// if we're not the leader, receiver receives entry-s from the leader
// and adds it to its entry log.
// if we're the leader, we also receive entry-s from self, and add it to our entry log

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func InitializeReceiver(port int) {
	// create an http server
	// which will listen on path /tokens/receive
	// and will receive tokens from the leader

	var log logger
	log = fmtLogger{}

	// create a channel to receive tokens from the leader
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		ctx := WithLogger(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	r.POST("/token/receive", receiveToken)
	// expose an endpoint to print all tokens in the entry log
	r.GET("/token/all", getAllTokens)
	r.Run(":" + strconv.Itoa(port))
}

func receiveToken(c *gin.Context) {
	log := GetLogger(c.Request.Context())
	// receive token from the leader
	// add token to the entry log
	// send token to the consumer

	var token string
	err := c.BindJSON(&token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid token",
		})
		return
	}

	log.Debugf("received token - %s", token)

	err = AddToEntryLog(token)
	if err != nil {
		fmt.Printf("error adding token to entry log - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "error adding token to entry log",
		})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func getAllTokens(c *gin.Context) {
	// get all tokens from the entry log
	// and send it to the consumer
	tokens := GetAllEntryLog()
	c.JSON(http.StatusOK, tokens)
}
