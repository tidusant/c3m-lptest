package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tidusant/c3m/common/c3mcommon"
	"github.com/tidusant/c3m/common/log"
	"github.com/tidusant/c3m/repo/models"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	//"github.com/gin-gonic/contrib/static"
)

var (
	loaddatadone         bool
	layoutPath           = "./template/out"
	schemePath           = "/scheme"
	schemeFolder         = "./scheme"
	rootFolder           = "./templates"
	rootPath             = "/templates"
	apiserver            string
	lptemplatetestserver string
)

func main() {
	initdata()
	if !loaddatadone {
		log.Errorf("Load data fail.")
		return
	}
	var port int
	var debug bool

	//check port
	rand.Seed(time.Now().Unix())
	port = 8082

	//fmt.Println(mycrypto.Encode("abc,efc", 5))
	flag.BoolVar(&debug, "debug", false, "Indicates if debug messages should be printed in log files")
	flag.Parse()

	logLevel := log.DebugLevel
	if !debug {
		layoutPath = "./layout"
		logLevel = log.InfoLevel
		gin.SetMode(gin.ReleaseMode)
		log.SetOutputFile(fmt.Sprintf("portal-"+strconv.Itoa(port)), logLevel)
		defer log.CloseOutputFile()
		log.RedirectStdOut()
	}
	log.Infof("debug %v", debug)

	//init config
	router := gin.Default()

	//http.Handle("/template/",  http.FileServer(http.Dir("./public")))

	//router.POST("/gettemplate", func(c *gin.Context) {
	//	rs:=HandleGetTemplate(c)
	//	b,_:=json.Marshal(rs)
	//	c.String(http.StatusOK, string(b))
	//})

	//router.Use(static.Serve("/", static.LocalFile("static", false)))
	router.StaticFile("/", layoutPath+"/index.html")
	//nextjs request File
	router.Static("/templates", "./templates")
	router.Static("/scheme", "./scheme")
	//router.StaticFile("/edit", layoutPath+"/edit.html")
	//router.LoadHTMLGlob(layoutPath+"/edit.html")

	router.GET("/test/:action/:params", HandleTestRoute)
	router.POST("/test/:action/:params", HandleTestRoute)

	log.Infof("running with port:" + strconv.Itoa(port))
	router.Run(":" + strconv.Itoa(port))

}

func HandleTestRoute(c *gin.Context) {
	//get cookie
	sex, _ := c.Cookie("_s")
	log.Debugf("cookies: %+v", sex)
	c.Writer.WriteHeader(http.StatusOK)
	if sex == "" {
		c.Writer.WriteString("Please login.")
		return
	}
	//get session to auth
	bodystr := c3mcommon.RequestAPI(apiserver, "aut", sex+"|t")
	var rs models.RequestResult
	err := json.Unmarshal([]byte(bodystr), &rs)

	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}
	if rs.Status != 1 {
		c.Writer.WriteString(rs.Error)
		return
	}
	var rt map[string]string
	err = json.Unmarshal([]byte(rs.Data), &rt)
	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}
	if v, ok := rt["username"]; !ok || v == "" {
		c.Writer.WriteString(`Please login again.`)
		return
	}
	//get modules permission from session
	modules := make(map[string]bool)
	for _, v := range strings.Split(rt["modules"], ",") {
		modules[v] = true
	}
	//check module permission
	if ok, _ := modules["c3m-lptpl-admin"]; !ok {
		c.Writer.WriteString(`Permission denied.`)
		return
	}

	action := c.Param("action")
	switch action {
	case "edit":
		c.Writer.WriteString(GetTest(c))
	case "submit":
		c.Writer.WriteString(SubmitTest(c))
	}

}

func initdata() {
	apiserver = os.Getenv("API_ADD")
	lptemplatetestserver = os.Getenv("LPTPLTEST_ADD")
	if len(apiserver) < 10 {
		log.Error("Api ip INVALID")
		os.Exit(0)
	}
	log.Printf("check version...")

	log.Printf("load layout...")

	loaddatadone = true
}
